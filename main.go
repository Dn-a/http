package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const SERVER_PORT = 3030
const BUFFER_SIZE = 30
const TIMEOUT = 10

func main() {

	/* f := getFile("m.txt")
	defer f.Close() */

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", SERVER_PORT))

	if err != nil {
		log.Fatal("Listener error: ", err)
	}
	fmt.Printf("Server started on %d\n", SERVER_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		func(c net.Conn) {
			defer c.Close()
			for line := range getLinesChannel(c) {
				fmt.Printf("%s", line)
			}
		}(conn)

		// ANSI escape sequence to clear terminal
		fmt.Print("\033[2J\033[H")
	}

}

func getFile(fName string) *os.File {
	f, err := os.Open(fName)
	if err != nil {
		log.Fatal("error", "error", err)
	}
	return f
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer f.Close()
		defer close(out)

		buffer := make([]byte, BUFFER_SIZE)
		lnIdx := -1
		for {
			_, err := f.Read(buffer)
			if err != nil {
				break
			}

			if lnIdx = bytes.IndexByte(buffer, '\n'); lnIdx != -1 {
				out <- string(buffer[:lnIdx+1])
			}

			if len(buffer[lnIdx+1:]) > 0 {
				out <- string(buffer[lnIdx+1:])
			}
			buffer = make([]byte, BUFFER_SIZE)
		}
	}()
	return out
}
