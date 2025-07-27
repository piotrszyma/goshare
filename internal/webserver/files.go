package webserver

import (
	"fmt"
	"os"
)

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
