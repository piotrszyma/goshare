package webserver

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	qrcode "github.com/skip2/go-qrcode"
)

// getUniqueFilename generates a unique filename by appending a suffix (.0, .1, etc.)
// if a file with the same name already exists in the specified directory
func getUniqueFilename(dir, filename string) string {
	// Check if the original file exists
	fullPath := filepath.Join(dir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// File doesn't exist, return original filename
		return filename
	}

	// File exists, try appending suffixes
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]

	counter := 0
	for {
		newFilename := fmt.Sprintf("%s.%d%s", name, counter, ext)
		newFullPath := filepath.Join(dir, newFilename)
		if _, err := os.Stat(newFullPath); os.IsNotExist(err) {
			// Found a unique filename
			return newFilename
		}
		counter++
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
