package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

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
	requestLine := strings.Split(request, CRLF)[0]

	fmt.Printf("Request line: %s", requestLine)
	url := strings.Split(requestLine, " ")[1]
	if url == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	} else if strings.HasPrefix(url, "/echo/") {
		str, _ := strings.CutPrefix(url, "/echo/")
		resp := fmt.Sprintf("HTTP/1.1 200 OK%sContent-Type: text/plain%sContent-Length: %d%s%s%s", CRLF, CRLF, len(str), CRLF, CRLF, str)
		conn.Write([]byte(resp))
		return
	}
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}
