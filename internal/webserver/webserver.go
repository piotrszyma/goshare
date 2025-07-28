package webserver

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

//go:embed templates/index.html
var indexHTML string

// templateData holds the data for the index template
type templateData struct {
	Message      string
	MessageType  string
	Key          string
	SharedFiles  []fileInfo
	UploadsFiles []fileInfo
}

// renderIndexTemplate renders the index.html template with the provided data
func renderIndexTemplate(w http.ResponseWriter, r *http.Request, uploadsDir, sharePath string) error {
	// Parse the embedded template
	tmpl, err := template.New("index.html").Parse(indexHTML)
	if err != nil {
		return err
	}

	// Prepare template data
	data := templateData{
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
	return tmpl.Execute(w, data)
}

// Run starts an HTTP server on the specified port that responds with a file upload form on the root path
// and handles file uploads on the /upload path
func Run(sharePath string, uploadsDir string, port int) {
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

		// Render the template with the appropriate data
		err := renderIndexTemplate(w, r, uploadsDir, sharePath)
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

	// Determine the port to use
	var actualPort int
	if port > 0 {
		// Validate port range
		if port < 1 || port > 65535 {
			log.Fatalf("Invalid port number: %d. Port must be between 1 and 65535.", port)
		}
		actualPort = port
	} else {
		// Find an available port
		listener, err := net.Listen("tcp", "0.0.0.0:0")
		if err != nil {
			log.Fatal("Failed to find available port:", err)
		}

		// Get the actual port that was assigned
		actualPort = listener.Addr().(*net.TCPAddr).Port

		// Close the listener temporarily - we'll recreate it with the http server
		listener.Close()
	}

	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("Warning: Could not determine local IP address: %v", err)
		log.Printf("Starting server on 0.0.0.0:%d", actualPort)
		fmt.Printf("Server URL: http://localhost:%d\n", actualPort)
	} else {
		serverURL := fmt.Sprintf("http://%s:%d", localIP, actualPort)
		serverURLWithKey := fmt.Sprintf("%s?key=%s", serverURL, secretKey)
		log.Printf("Starting server on 0.0.0.0:%d (accessible from: %s)", actualPort, serverURLWithKey)
		fmt.Printf("Server URL: %s\n", serverURLWithKey)

		// Print QR code for easy mobile access
		printQRCode(serverURLWithKey)
	}

	address := fmt.Sprintf("0.0.0.0:%d", actualPort)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal(err)
	}
}
