package tritonhttp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// DocRoot specifies the path to the directory to serve static files from.
	DocRoot string
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {
	//panic("todo")
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		fmt.Println(" Could not open Listener")
		return err
	}
	defer ln.Close()
	// Hint: call HandleConnection

	for {

		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Could not receive a connection, thus going back to accept")
			continue
		}

		fmt.Println("Connected to client at:", conn.RemoteAddr())
		go s.HandleConnection(conn)
	}

}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	//panic("todo")

	// Hint: use the other methods below
	br := bufio.NewReader(conn)
	defer conn.Close()
	for {
		// Set timeout
		err := conn.SetReadDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			fmt.Println("Could not set deadline")
			return
		}
		// Try to read next request

		req, bytesReceived, err := ReadRequest(br)
		// Handle EOF
		// Handle timeout
		if err == io.EOF {
			fmt.Println("Connection closed by client or request ended")
			return
		}

		if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
			fmt.Printf("Connection to %v timed out", conn.RemoteAddr())
			if bytesReceived == true {
				fmt.Println(" Partial req had been recieved so sending 400 back")
				res := &Response{}
				res.HandleBadRequest()
				wErr := res.Write(conn)
				if wErr != nil {
					fmt.Println("Error in writing bad response after timeout")
				}
			}
			return
		}

		// Handle bad request 400 {Handle header with missing colon, incomplete request}
		if err != nil {
			fmt.Println("Recieved a malformed request, sending a bad response")
			res := &Response{}
			res.HandleBadRequest()
			wErr := res.Write(conn)
			if wErr != nil {
				fmt.Println("Error in writing bad response after timeout")
			}
			return
		}
		// Handle good request
		res := s.HandleGoodRequest(req)
		wErr := res.Write(conn)
		if wErr != nil {
			fmt.Println("Error in writing good request response")
		}
		// Close conn if requested
		if res.Request.Close == true {
			fmt.Println("Closing connection as requested by client")
			return
		} else {
			continue
		}
	}
}

// HandleGoodRequest handles the valid req and generates the corresponding res.
func (s *Server) HandleGoodRequest(req *Request) (res *Response) {
	//panic("todo")
	var serverRes Response

	// Hint: use the other methods below
	reqURL := req.URL
	//fmt.Println("After req parsing:", reqURL)
	if reqURL[0] != '/' || req.Host == "" {
		serverRes.HandleBadRequest()
		return &serverRes
	}
	if reqURL[len(reqURL)-1] == '/' {
		reqURL += "index.html"
	}
	//fmt.Println("After adding file path", reqURL)
	fileLoc := filepath.Join(s.DocRoot, reqURL)

	// Revisit this
	if strings.HasPrefix(fileLoc, s.DocRoot) == false {
		serverRes.HandleNotFound(req)
		return &serverRes
	}

	_, err := os.Stat(fileLoc)

	if os.IsNotExist(err) {
		serverRes.HandleNotFound(req)
		return &serverRes
	}

	serverRes.HandleOK(req, fileLoc)
	return &serverRes
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK(req *Request, path string) {
	//panic("todo")
	header := make(map[string]string)
	res.StatusCode = 200
	res.Proto = "HTTP/1.1"
	res.Request = req
	res.FilePath = path

	fileStat, _ := os.Stat(path)

	header["Date"] = FormatTime(time.Now())

	header[CanonicalHeaderKey("last-modified")] = FormatTime(fileStat.ModTime())

	ext := strings.SplitAfter(path, ".")
	header[CanonicalHeaderKey("content-type")] = MIMETypeByExtension("." + ext[len(ext)-1])

	header[CanonicalHeaderKey("content-length")] = strconv.Itoa(int(fileStat.Size()))

	if req.Close {
		header["Connection"] = "close"
	}
	res.Header = header
}

// HandleBadRequest prepares res to be a 400 Bad Request response
// ready to be written back to client.
func (res *Response) HandleBadRequest() {
	//panic("todo")
	header := make(map[string]string)
	res.StatusCode = 400
	res.Proto = "HTTP/1.1"

	header["Connection"] = "close"
	header["Date"] = FormatTime(time.Now())
	res.Header = header
}

// HandleNotFound prepares res to be a 404 Not Found response
// ready to be written back to client.
func (res *Response) HandleNotFound(req *Request) {
	//panic("todo")
	header := make(map[string]string)
	res.StatusCode = 404
	res.Proto = "HTTP/1.1"
	res.Request = req

	header["Date"] = FormatTime(time.Now())
	if req.Close {
		header["Connection"] = "close"
	}
	res.Header = header
}
