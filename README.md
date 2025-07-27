-----

# GoShare: Simple File Sharing over Local Network

GoShare is a lightweight Go application that turns your machine into a local file server, enabling seamless file uploads and downloads within your local network. It's perfect for quickly sharing files between devices without relying on external services.

## Features

  * **Easy File Uploads:** Clients can easily upload files to the server's current working directory via a simple web interface.
  * **Convenient File Downloads:** Share specific files from your server, making them accessible for download by other devices on your network.
  * **Web Interface:** A user-friendly web interface at the root URL allows clients to list shared files and upload new ones.
  * **Command-Line Simplicity:** Run the server with minimal configuration directly from your terminal.

## Project Naming Convention

For Go projects that are executable applications, especially command-line tools or servers, it's common practice to name the project after the main binary or its primary function. Given GoShare's purpose, **GoShare** is an intuitive and descriptive name. It clearly indicates it's a Go application for sharing.

## How to Use

### Installation

To get started, make sure you have Go installed on your system. Then, you can install GoShare by running:

```bash
go install github.com/your-username/goshare@latest # Replace with your actual repository path
```

This will install the `goshare` executable in your `$GOPATH/bin` directory (or `$GOBIN` if set), making it available from your command line.

### Running the Server

Here are some examples of how to run the GoShare server:

#### Basic Server (Uploads Only)

To run the server and allow clients to upload files to your current working directory (where you execute the command):

```bash
goshare run
```

This will start the server on port 8000, accessible from other devices on your local network. You can access the web interface from other devices by navigating to `http://<your-server-ip>:8000` in a web browser.

When the server starts, it will print a QR code to the console that you can scan with your mobile device to easily access the file sharing interface.

#### Sharing Specific Files (Uploads and Downloads)

To run the server and also allow clients to download a specific file (e.g., `./my_document.pdf`):

```bash
goshare run --share ./my_document.pdf
```

You can share multiple files by repeating the `--share` flag:

```bash
goshare run --share ./my_document.pdf --share ./my_image.jpg
```

-----

## API Examples

### `goshare run`

This command initializes the GoShare server. When executed without the `--share` flag, it primarily functions as an upload server. The web interface will allow clients to upload files to the directory from which `goshare` was launched.

### `goshare run --share <filepath>`

This command extends the server's functionality to include file downloads. The specified `<filepath>` will be made available for clients to download via the web interface. Clients will see a list of shared files on the root HTML page, along with the upload functionality.

-----

## Server Web Interface

The local server will expose a single HTML page at its root (`/`). This page will provide the following functionalities:

1.  **Shared File Listing and Download:** If files were shared using the `--share` flag, they will be listed on this page, and clients will be able to click on them to initiate downloads.
2.  **File Upload Form:** A form will be present on the page, allowing clients to select and upload files from their local machine to the server's current working directory.

-----

## Contributing

We welcome contributions to GoShare\! Feel free to open issues or submit pull requests on the GitHub repository.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
