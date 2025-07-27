package webserver

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

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

		// Generate a unique filename if file already exists
		uniqueFilename := getUniqueFilename(uploadsDir, handler.Filename)

		// Create destination file with unique name
		dst, err := os.Create(fmt.Sprintf("%s/%s", uploadsDir, uniqueFilename))
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

	// Find an available port
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		log.Fatal("Failed to find available port:", err)
	}

	// Get the actual port that was assigned
	port := listener.Addr().(*net.TCPAddr).Port

	// Close the listener temporarily - we'll recreate it with the http server
	listener.Close()

	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("Warning: Could not determine local IP address: %v", err)
		log.Printf("Starting server on 0.0.0.0:%d", port)
		fmt.Printf("Server URL: http://localhost:%d\n", port)
	} else {
		serverURL := fmt.Sprintf("http://%s:%d", localIP, port)
		serverURLWithKey := fmt.Sprintf("%s?key=%s", serverURL, secretKey)
		log.Printf("Starting server on 0.0.0.0:%d (accessible from: %s)", port, serverURLWithKey)
		fmt.Printf("Server URL: %s\n", serverURLWithKey)

		// Print QR code for easy mobile access
		printQRCode(serverURLWithKey)
	}

	address := fmt.Sprintf("0.0.0.0:%d", port)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
