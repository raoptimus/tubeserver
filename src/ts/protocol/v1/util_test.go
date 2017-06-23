package v1

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"
)

// to generate example output run go test with -v flag
func runExample(request string) string {
	conn := getRpcConn()
	defer conn.Close()
	_, err := conn.Write([]byte(request))
	if err != nil {
		log.Fatalf("Request %s return error %v", request, err)
	}
	buf := make([]byte, 24*1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Request %s return error %v", request, err)
	}
	response := string(buf[:n])

	// generate docs
	if testing.Verbose() {
		indent := func(header, data string) {
			print("### " + header + ":\n")
			var dst bytes.Buffer
			if err := json.Indent(&dst, []byte(data), "", "  "); err != nil {
				log.Fatalf("%s json unmarshal error %v", header, err)
			}
			print(dst.String() + "\n")
		}
		indent("Request", request)
		indent("Response", response)
	}
	return response
}
