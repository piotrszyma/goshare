-----

# 🚀 GoShare: Simple File Sharing over Local Network

GoShare is a lightweight Go application that turns your machine into a local file server, enabling seamless file uploads and downloads within your local network. It's perfect for quickly sharing files between devices without relying on external services.

Built with security in mind, GoShare uses a secret key authentication system to ensure only authorized users can access your files.

## 🌟 Features

  * **📤 Easy File Uploads:** Clients can easily upload files to the server's current working directory via a simple web interface.
  * **📥 Convenient File Downloads:** Share specific files from your server, making them accessible for download by other devices on your network.
  * **🌐 Web Interface:** A user-friendly web interface at the root URL allows clients to list shared files and upload new ones.
  * **💻 Command-Line Simplicity:** Run the server with minimal configuration directly from your terminal.

## 💡 Project Naming Convention

For Go projects that are executable applications, especially command-line tools or servers, it's common practice to name the project after the main binary or its primary function. Given GoShare's purpose, **GoShare** is an intuitive and descriptive name. It clearly indicates it's a Go application for sharing.

## 📖 How to Use

### ⚙️ Installation

To get started, make sure you have Go installed on your system. Then, you can install GoShare by running:

```bash
go install github.com/piotrszyma/goshare@latest
```

This will install the `goshare` executable in your `$GOPATH/bin` directory (or `$GOBIN` if set), making it available from your command line.

### 🚀 Running the Server

Here are some examples of how to run the GoShare server:

#### 📤 Basic Server (Uploads Only)

To run the server and allow clients to upload files to your current working directory (where you execute the command):

```bash
goshare
```

This will start the server on port 8000, accessible from other devices on your local network. You can access the web interface from other devices by navigating to `http://<your-server-ip>:8000` in a web browser.

When the server starts, it will print a QR code to the console that you can scan with your mobile device to easily access the file sharing interface.

#### 📥 Sharing Specific Files (Uploads and Downloads)

To run the server and also allow clients to download a specific file (e.g., `./my_document.pdf`):

```bash
goshare --share ./my_document.pdf
```

You can share multiple files by repeating the `--share` flag:

```bash
goshare --share ./my_document.pdf --share ./my_image.jpg
```

#### 📁 Specifying Upload Directory

By default, uploaded files are stored in an `uploads/` directory. You can specify a different directory using the `--uploads-dir` flag:

```bash
goshare --uploads-dir ./my-uploads
```

#### 🔍 Checking Version

To check the version of GoShare:

```bash
goshare version
```

-----

## 🧪 API Examples

### `goshare`

This command initializes the GoShare server. When executed without the `--share` flag, it primarily functions as an upload server. The web interface will allow clients to upload files to the directory from which `goshare` was launched.

### `goshare --share <filepath>`

This command extends the server's functionality to include file downloads. The specified `<filepath>` will be made available for clients to download via the web interface. Clients will see a list of shared files on the root HTML page, along with the upload functionality.

### `goshare --uploads-dir <directory>`

This command specifies a custom directory for storing uploaded files. By default, files are stored in an `uploads/` directory.

### `goshare version`

This command displays the current version of GoShare.

-----

## 🌐 Server Web Interface

The local server exposes a single HTML page at its root (`/`) with a clean, responsive design. This page provides the following functionalities:

1.  **Authentication:** Access to the web interface requires a secret key that is generated when the server starts. This key is automatically applied when accessing the server URL printed to the console, or can be manually added as a `key` parameter in the URL.

2.  **Shared File Listing and Download:** If files were shared using the `--share` flag, they will be listed on this page with their file sizes, and clients can click on them to initiate downloads.

3.  **Uploaded File Listing:** Files that have been uploaded to the server are displayed in a separate section with their file sizes and download links.

4.  **File Upload Form:** A form allows clients to select and upload files from their local machine to the server's upload directory.

5.  **QR Code Access:** When the server starts, a QR code is printed to the console that can be scanned with a mobile device to easily access the file sharing interface with the required authentication key.

## 🔒 Security Features

GoShare implements several security measures to protect your files:

- **Secret Key Authentication:** A randomly generated secret key is required to access the web interface, preventing unauthorized access to your files.

- **Secure Cookie Handling:** After initial authentication, a secure cookie is used to maintain the session, with automatic expiration after 1 hour.

- **Path Restriction:** File access is restricted to only the shared files and upload directory, preventing access to other parts of the filesystem.

- **Request Logging:** All requests are logged with timestamps and response codes for monitoring access to your server.

-----

## 🤝 Contributing

We welcome contributions to GoShare\! Feel free to open issues or submit pull requests on the GitHub repository.

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.
