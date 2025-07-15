package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type StatusLine struct {
	version    string
	statusCode int
	reason     string
}

type RequestLine struct {
	httpMethod    string
	requestTarget string
	httpVersion   string
}

type HTTPResponse struct {
	statusLine StatusLine
	headers    map[string]string
	body       string
}

type HTTPRequest struct {
	requestLine RequestLine
	headers     map[string]string
	body        string
}

func main() {
	directoryPtr := flag.String("directory", "/tmp/", "The directory the server will look for files")
	flag.Parse()
	directory := *directoryPtr

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	fmt.Printf("Server started. Directory: %s\n", directory)

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(conn, directory)

	}
}

func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()

	const CRLF = "\r\n"

	buff := make([]byte, 1024)
	conn.Read(buff)

	request := string(buff)
	splitRequest := strings.Split(request, CRLF)
	splitRequestLine := strings.Split(splitRequest[0], " ")
	headers := extractHeaders(splitRequest[1:])
	body := ""
	if _, ok := headers["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(headers["Content-Length"])
		if err != nil {
			fmt.Printf("Error converting content length from header to int")
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		if splitRequest[len(splitRequest)-2] == "" {
			body = splitRequest[len(splitRequest)-1][:contentLength]
		}
	}
	httpReq := HTTPRequest{
		requestLine: RequestLine{
			httpMethod:    splitRequestLine[0],
			requestTarget: splitRequestLine[1],
			httpVersion:   splitRequestLine[2],
		},
		headers: headers,
		body:    body,
	}

	fmt.Printf("Request body string len: %d\nRequest body string: %s\n", len(body), body)
	url := httpReq.requestLine.requestTarget
	if url == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	} else if strings.HasPrefix(url, "/echo/") {
		str, _ := strings.CutPrefix(url, "/echo/")
		resp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(str), str)
		conn.Write([]byte(resp))
		return
	} else if url == "/user-agent" {
		for _, line := range splitRequest {
			if strings.HasPrefix(strings.ToLower(line), "user-agent: ") {
				userAgent := line[12:]
				resp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
				conn.Write([]byte(resp))
				return
			}
		}
	} else if strings.HasPrefix(url, "/files/") && httpReq.requestLine.httpMethod == "GET" {
		fileName, _ := strings.CutPrefix(url, "/files/")
		content, err := os.ReadFile(fmt.Sprintf("%s%s", directory, fileName))
		if err != nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)))
		return
	} else if strings.HasPrefix(url, "/files/") && httpReq.requestLine.httpMethod == "POST" {
		fileName, _ := strings.CutPrefix(url, "/files/")
		fullPath := filepath.Join(directory, fileName)
		file, err := os.Create(fullPath)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			fmt.Printf("Failed to create file: %v", err)
			return
		}
		defer file.Close()
		fmt.Println("Successfully created file.")
		n, err := file.WriteString(httpReq.body)
		if err != nil {
			fmt.Printf("Error writing to file: %v", err)
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		fmt.Printf("Wrote %d bytes into file", n)
		conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		return
	}
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func extractHeaders(splitRequest []string) map[string]string {
	headers := make(map[string]string)
	for _, line := range splitRequest {
		if line == "" {
			return headers
		}
		splitHeader := strings.Split(line, ": ")
		headers[splitHeader[0]] = splitHeader[1]
	}
	return headers
}
