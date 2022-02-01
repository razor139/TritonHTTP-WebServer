package tritonhttp

import (
	"bufio"
	"errors"
	"regexp"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Header stores misc headers excluding "Host" and "Connection",
	// which are stored in special fields below.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

// ReadRequest tries to read the next valid request from br.
//
// If it succeeds, it returns the valid request read. In this case,
// bytesReceived should be true, and err should be nil.
//
// If an error occurs during the reading, it returns the error,
// and a nil request. In this case, bytesReceived indicates whether or not
// some bytes are received before the error occurs. This is useful to determine
// the timeout with partial request received condition.
func ReadRequest(br *bufio.Reader) (req *Request, bytesReceived bool, err error) {
	//panic("todo")
	req = &Request{}
	var assignHost int = 0
	Header := make(map[string]string)

	// Read start line
	line, err := ReadLine(br)
	if err != nil {
		return nil, (len(line) != 0), err
	}

	firstLine := strings.Split(line, " ")
	if len(firstLine) != 3 {
		return nil, false, errors.New("Initial request line malformed")
	}

	if firstLine[0] == "GET" {
		req.Method = "GET"
	} else {
		return nil, false, errors.New("Method is not GET")
	}

	if strings.HasPrefix(firstLine[1], "/") {
		req.URL = firstLine[1]
	} else {
		return nil, false, errors.New(" URL not starting with /")
	}

	if firstLine[2] == "HTTP/1.1" {
		req.Proto = firstLine[2]
	} else {
		return nil, false, errors.New(" Proto is not correct ")
	}

	// Reading request headers
	for {
		line, err := ReadLine(br)
		//fmt.Println(len(line), " ", line)
		if err != nil {
			return req, true, err
		}
		if len(line) == 0 {
			//fmt.Println(line, " Reached end of 1st req")
			if assignHost == 1 {
				break
			} else {
				return nil, false, errors.New("Bad CLRF in middle of request")
			}

		}

		headerRegex := regexp.MustCompile(`^([a-zA-Z0-9-]+):[ ]*(.*[\r]*[\n]*.*)$`)
		splitLine := headerRegex.FindStringSubmatch(line)
		//fmt.Println("Line split:", splitLine)
		if len(splitLine) == 3 {
			key := splitLine[1]
			if strings.EqualFold(key, "Host") {
				if splitLine[2] == "" {
					req.Host = "default"
				} else {
					req.Host = splitLine[2]
				}
				assignHost = 1
			} else if strings.EqualFold(key, "Connection") {
				if splitLine[2] == "close" {
					req.Close = true
				} else {
					req.Close = false
				}
			} else {
				keyRegex := regexp.MustCompile("^[a-zA-Z0-9-]*$")
				if keyRegex.MatchString(key) {
					Header[CanonicalHeaderKey(key)] = splitLine[2]
				} else {
					return nil, false, errors.New("Key is not alphanumeric with hyphen")
				}
			}

		} else {
			return nil, false, errors.New("Header is malformed")
		}
	}

	// Read headers
	req.Header = Header
	// Check required headers
	// Handle special headers

	return req, true, nil
}
