package webserver

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// responseWriter is a wrapper around http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader overrides the http.ResponseWriter's WriteHeader method to capture the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs each request and response status
func loggingMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Record the start time
		start := time.Now()

		// Wrap the ResponseWriter to capture the status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		handler(wrapped, r)

		// Log the request details
		log.Printf(
			"%s %s %s %d %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
		)
	}
}

var secretKey string

func init() {
	// Generate a random secret key at startup
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Failed to generate random secret key:", err)
	}
	secretKey = hex.EncodeToString(bytes)
}

// validateKey checks if the request has a valid key parameter
func validateKey(r *http.Request) bool {
	keys, ok := r.URL.Query()["key"]

	log.Printf("request with keys = %s", keys)

	if !ok || len(keys) == 0 {
		return false
	}
	return keys[0] == secretKey
}

// validateKeyCookie checks if the request has a valid key cookie
func validateKeyCookie(r *http.Request) bool {
	cookie, err := r.Cookie("key")
	if err != nil {
		return false
	}

	log.Printf("request with key cookie = %s", cookie.Value)

	return cookie.Value == secretKey
}

// requireKey is middleware that checks for a valid key cookie
func requireKey(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateKeyCookie(r) {
			http.Error(w, "Unauthorized: invalid or missing key cookie", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

// getLocalIP returns the non-loopback local IP of the host
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no local IP address found")
}

// printQRCode prints a QR code to the console for the given URL
func printQRCode(url string) {
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		log.Printf("Failed to generate QR code: %v", err)
		return
	}

	fmt.Println("\nScan this QR code with your mobile device to access the file sharing server:")
	fmt.Println(qr.ToSmallString(false))
}

// fileInfo represents information about a file for display in the UI
type fileInfo struct {
	Name string
	Size int64
	URL  string
}

// getSharedFiles returns a list of files to be shared based on the provided path
func getSharedFiles(sharePath string) ([]fileInfo, error) {
	var files []fileInfo

	// Check if the path exists
	info, err := os.Stat(sharePath)
	if err != nil {
		return nil, err
	}

	// If it's a file, add just that file
	if !info.IsDir() {
		files = append(files, fileInfo{
			Name: info.Name(),
			Size: info.Size(),
			URL:  "/shared/" + info.Name(),
		})
		return files, nil
	}

	// If it's a directory, add all files in the directory (not subdirectories)
	entries, err := os.ReadDir(sharePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Skip subdirectories
		if entry.IsDir() {
			continue
		}

		// Get file info
		fileInfoStat, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, fileInfo{
			Name: fileInfoStat.Name(),
			Size: fileInfoStat.Size(),
			URL:  "/shared/" + fileInfoStat.Name(),
		})
	}

	return files, nil
}

// getUploadsFiles returns a list of files in the uploads directory
func getUploadsFiles(uploadsDir string) ([]fileInfo, error) {
	var files []fileInfo

	// Check if the uploads directory exists
	info, err := os.Stat(uploadsDir)
	if err != nil {
		// If directory doesn't exist, return empty list
		return files, nil
	}

	// If it's not a directory, return error
	if !info.IsDir() {
		return nil, fmt.Errorf("uploads path is not a directory")
	}

	// Read files in the directory
	entries, err := os.ReadDir(uploadsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Skip subdirectories
		if entry.IsDir() {
			continue
		}

		// Get file info
		fileInfoStat, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, fileInfo{
			Name: fileInfoStat.Name(),
			Size: fileInfoStat.Size(),
			URL:  "/uploads/" + fileInfoStat.Name(),
		})
	}

	return files, nil
}

// Run starts an HTTP server on port 8000 that responds with a file upload form on the root path
// and handles file uploads on the /upload path
func Run(sharePath string, uploadsDir string) {
	// Set default uploads directory if not provided
	defaultUploadsDir := "uploads"
	if uploadsDir == "" {
		uploadsDir = defaultUploadsDir
	} else {
		// Check if specified uploads directory already exists
		if _, err := os.Stat(uploadsDir); err == nil {
			log.Fatalf("Error: uploads directory '%s' already exists in current working directory", uploadsDir)
		}

		// Create specified uploads directory
		err := os.MkdirAll(uploadsDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Error creating uploads directory: %v", err)
		}
	}
	// If sharePath is provided, set up file serving
	if sharePath != "" {
		// Check if the path exists
		_, err := os.Stat(sharePath)
		if err == nil {
			// Serve files from the shared path
			// If it's a file, serve it directly
			// If it's a directory, serve files from that directory
			info, _ := os.Stat(sharePath)
			if info.IsDir() {
				// Serve files from the directory
				fileServer := http.StripPrefix("/shared/", http.FileServer(http.Dir(sharePath)))
				http.Handle("/shared/", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
					// Apply authentication check
					if !validateKeyCookie(r) {
						http.Error(w, "Unauthorized: invalid or missing key cookie", http.StatusUnauthorized)
						return
					}
					fileServer.ServeHTTP(w, r)
				}))
			} else {
				// Serve the single file
				http.HandleFunc("/shared/"+info.Name(), loggingMiddleware(requireKey(func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, sharePath)
				})))
			}
		}
	}
	// Set up file serving for uploads directory
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir)))
	http.Handle("/uploads/", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Apply authentication check
		if !validateKeyCookie(r) {
			http.Error(w, "Unauthorized: invalid or missing key cookie", http.StatusUnauthorized)
			return
		}
		fileServer.ServeHTTP(w, r)
	}))

	// Handle root path - serve HTML with file upload form and shared files
	http.HandleFunc("/", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {

		if !validateKeyCookie((r)) {
			if validateKey(r) {

				// Set the key as a cookie with enhanced security
				http.SetCookie(w, &http.Cookie{
					Name:     "key",
					Value:    secretKey,
					Path:     "/",
					HttpOnly: true,
					// Secure: true,
					// SameSite: http.SameSiteStrictMode,
					MaxAge: 3600, // 1 hour
				})

				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			} else {
				http.Error(w, "No key provided", http.StatusForbidden)
				return
			}
		}

		// Define the HTML template with file upload form and shared files
		htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoShare - File Sharing</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
        }
        h2 {
            color: #555;
            margin-top: 30px;
        }
        form {
            display: flex;
            flex-direction: column;
            gap: 15px;
        }
        input[type="file"] {
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
        }
        input[type="submit"] {
            padding: 12px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
        }
        input[type="submit"]:hover {
            background-color: #0056b3;
        }
        .message {
            padding: 15px;
            margin: 15px 0;
            border-radius: 5px;
        }
        .success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .file-list {
            list-style-type: none;
            padding: 0;
        }
        .file-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px;
            border-bottom: 1px solid #eee;
        }
        .file-item:last-child {
            border-bottom: none;
        }
        .file-link {
            color: #007bff;
            text-decoration: none;
        }
        .file-link:hover {
            text-decoration: underline;
        }
        .file-size {
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>GoShare File Sharing</h1>

        {{if .Message}}
            <div class="message {{.MessageType}}">
                {{.Message}}
            </div>
        {{end}}

        {{if .UploadsFiles}}
        <h2>Uploaded Files</h2>
        <ul class="file-list">
            {{range .UploadsFiles}}
            <li class="file-item">
                <a href="{{.URL}}" class="file-link">{{.Name}}</a>
                <span class="file-size">({{.Size}} bytes)</span>
            </li>
            {{end}}
        </ul>
        {{end}}

        {{if .SharedFiles}}
        <h2>Shared Files</h2>
        <ul class="file-list">
            {{range .SharedFiles}}
            <li class="file-item">
                <a href="{{.URL}}" class="file-link">{{.Name}}</a>
                <span class="file-size">({{.Size}} bytes)</span>
            </li>
            {{end}}
        </ul>
        {{end}}

        <h2>Upload New File</h2>
        <form action="/upload?key={{.Key}}" method="post" enctype="multipart/form-data">
            <input type="file" name="file" required>
            <input type="submit" value="Upload File">
        </form>
    </div>
</body>
</html>
`
		// Parse the template
		tmpl, err := template.New("index").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Prepare template data
		data := struct {
			Message      string
			MessageType  string
			Key          string
			SharedFiles  []fileInfo
			UploadsFiles []fileInfo
		}{
			Key: secretKey,
		}

		// Get files from uploads directory
		uploadsFileInfoList, err := getUploadsFiles(uploadsDir)
		if err != nil {
			data.Message = "Error accessing uploads directory: " + err.Error()
			data.MessageType = "error"
		} else {
			data.UploadsFiles = uploadsFileInfoList
		}

		// If sharePath is provided, get file info to display
		if sharePath != "" {
			// Check if it's a file or directory
			fileInfoList, err := getSharedFiles(sharePath)
			if err != nil {
				data.Message = "Error accessing shared path: " + err.Error()
				data.MessageType = "error"
			} else {
				data.SharedFiles = fileInfoList
			}
		}

		// Get any message from query parameters
		if message := r.URL.Query().Get("message"); message != "" {
			data.Message = message
			if messageType := r.URL.Query().Get("type"); messageType != "" {
				data.MessageType = messageType
			} else {
				data.MessageType = "success"
			}
		}

		// Execute the template
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}))

	// Handle file upload
	http.HandleFunc("/upload", loggingMiddleware(requireKey(func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form with max memory of 32MB
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			http.Redirect(w, r, "/?message="+err.Error()+"&type=error", http.StatusSeeOther)
			return
		}

		// Get the file from the form
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Redirect(w, r, "/?message=Error retrieving file: "+err.Error()+"&type=error", http.StatusSeeOther)
			return
		}
		defer file.Close()

		// Create uploads directory if it doesn't exist
		err = os.MkdirAll(uploadsDir, os.ModePerm)
		if err != nil {
			http.Redirect(w, r, "/?message=Error creating uploads directory: "+err.Error()+"&type=error", http.StatusSeeOther)
			return
		}

		// Create destination file
		dst, err := os.Create(fmt.Sprintf("%s/%s", uploadsDir, handler.Filename))
		if err != nil {
			http.Redirect(w, r, "/?message=Error creating file: "+err.Error()+"&type=error", http.StatusSeeOther)
			return
		}
		defer dst.Close()

		// Copy uploaded file to destination
		_, err = io.Copy(dst, file)
		if err != nil {
			http.Redirect(w, r, "/?message=Error saving file: "+err.Error()+"&type=error", http.StatusSeeOther)
			return
		}

		// Redirect back to home page with success message
		http.Redirect(w, r, "/?message=File uploaded successfully!&type=success", http.StatusSeeOther)
	})))

	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("Warning: Could not determine local IP address: %v", err)
		log.Println("Starting server on 0.0.0.0:8000")
		fmt.Println("Server URL: http://localhost:8000")
	} else {
		serverURL := fmt.Sprintf("http://%s:8000", localIP)
		serverURLWithKey := fmt.Sprintf("%s?key=%s", serverURL, secretKey)
		log.Printf("Starting server on 0.0.0.0:8000 (accessible from: %s)", serverURLWithKey)
		fmt.Printf("Server URL: %s\n", serverURLWithKey)

		// Print QR code for easy mobile access
		printQRCode(serverURLWithKey)
	}

	err = http.ListenAndServe("0.0.0.0:8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
