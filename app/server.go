package main

import (
	"fmt"
	"net"
	"os"
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
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(conn)

	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	const CRLF = "\r\n"

	buff := make([]byte, 1024)
	conn.Read(buff)

	request := string(buff)
	splitRequest := strings.Split(request, CRLF)
	splitRequestLine := strings.Split(splitRequest[0], " ")
	body := ""
	if splitRequest[len(splitRequest)-2] == "" {
		body = splitRequest[len(splitRequest)-1]
	}
	httpReq := HTTPRequest{
		requestLine: RequestLine{
			httpMethod:    splitRequestLine[0],
			requestTarget: splitRequestLine[1],
			httpVersion:   splitRequestLine[2],
		},
		headers: extractHeaders(splitRequest[1:]),
		body:    body,
	}

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
				resp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d%s", len(userAgent), userAgent)
				conn.Write([]byte(resp))
				return
			}
		}
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
