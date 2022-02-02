package tritonhttp

import (
	"bufio"
	"io"
	"os"
	"sort"
	"strconv"
)

type Response struct {
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.1"

	// Header stores all headers to write to the response.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

// Write writes the res to the w.
func (res *Response) Write(w io.Writer) error {
	if err := res.WriteStatusLine(w); err != nil {
		return err
	}
	if err := res.WriteSortedHeaders(w); err != nil {
		return err
	}
	if err := res.WriteBody(w); err != nil {
		return err
	}
	return nil
}

// WriteStatusLine writes the status line of res to w, including the ending "\r\n".
// For example, it could write "HTTP/1.1 200 OK\r\n".
func (res *Response) WriteStatusLine(w io.Writer) error {
	//panic("todo")
	statusLine := res.Proto + " " + strconv.Itoa(res.StatusCode) + " "
	if res.StatusCode == 200 {
		statusLine += "OK\r\n"
	} else if res.StatusCode == 400 {
		statusLine += "Bad Request\r\n"
	} else if res.StatusCode == 404 {
		statusLine += "Not Found\r\n"
	} else {
		return nil
	}

	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	return nil
}

// WriteSortedHeaders writes the headers of res to w, including the ending "\r\n".
// For example, it could write "Connection: close\r\nDate: foobar\r\n\r\n".
// For HTTP, there is no need to write headers in any particular order.
// TritonHTTP requires to write in sorted order for the ease of testing.
func (res *Response) WriteSortedHeaders(w io.Writer) error {
	//panic("todo")
	var header string
	keys := make([]string, 0, len(res.Header))
	for key := range res.Header {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		header += key + ":" + " " + res.Header[key] + "\r\n"
	}
	header += "\r\n"
	_, err := w.Write([]byte(header))
	if err != nil {
		return err
	}
	return nil
}

// WriteBody writes res' file content as the response body to w.
// It doesn't write anything if there is no file to serve.
func (res *Response) WriteBody(w io.Writer) error {
	//panic("todo")
	bw := bufio.NewWriter(w)

	if res.FilePath == "" {
		return nil
	}

	f, err := os.Open(res.FilePath)
	if err != nil {
		return err
	}

	//var buff1 []byte
	buff := make([]byte, 100)
	for {
		n, err := f.Read(buff)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		if _, err := bw.Write(buff[:n]); err != nil {
			return err
		}
		//buff1 = append(buff1, buff[:n]...)
	}

	if err := bw.Flush(); err != nil {
		return err
	}

	// _, err1 := w.Write(buff1)
	// if err1 != nil {
	// 	return err
	// }
	return nil
}
