// qrserve - HTTP microservice for QR Code generation
// Copyright 2016 Good Code

// You may use and/or distribute this software under the terms of MIT license
// See the README.md file for details

package main

// The service uses qrcode package from https://github.com/skip2/go-qrcode for
// QR code generation. This is the only dependency outside the Go standard
// library.
import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// We do have limits, and here they are - image can't be larger than 4k x 4k
const (
	MaxSize = 4096
)

// We only have one endpoint, /, and this is the handler function that should
// be called for each request to our endpoint.
func qrHandler(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.Path, "/")
	var data string
	var sizeStr string
	var q string
	partsLen := len(parts)
	if partsLen > 2 {
		data = parts[1]
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err == nil {
			data = string(decoded)
		}
		sizeStr = parts[2]
		if len(parts) > 3 {
			q = parts[3]
		}
	} else {
		// First we need to parse the query string so we can pick up the values
		if err := req.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data = req.FormValue("data")
		sizeStr = req.FormValue("size")
		q = req.FormValue("q")
	}
	q = strings.ToUpper(q)

	// Next we need to get the data (text to be encoded), size (of the generated
	// code, in pixels),  and error correction level (one of L, Q, M or H).
	if data == "" {
		http.Error(w, "Data must not be empty", http.StatusBadRequest)
		return
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		http.Error(w, "Error parsing size: "+err.Error(), http.StatusBadRequest)
		return
	}

	if size < 1 || size > MaxSize {
		http.Error(w, "Invalid image size: "+string(size), http.StatusBadRequest)
		return
	}

	level := qrcode.Medium // default
	switch q {
	case "L":
		level = qrcode.Low
	case "Q":
		level = qrcode.High
	case "H":
		level = qrcode.Highest
	}

	// Next we call the fine qrcode library to do the heavy lifting
	image, err := qrcode.Encode(data, level, size)
	if err != nil {
		http.Error(w, "Error creating QR code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Serve the resulting PNG image to the client and we're done!
	w.Header().Set("Content-Type", "image/png")
	_, err = w.Write(image)
	if err != nil {
		log.Printf("Error writing image: %v", err)
	}
}

func main() {
	var listenAddr string

	// Our command takes just one parameter, the address we should listen to.
	// If it is not given, we might as well tell the user what we expect.
	if len(os.Args) == 2 {
		listenAddr = os.Args[1]
	} else {
		fmt.Fprintf(os.Stderr, "Usage: %s [address]:port\n", os.Args[0])
		os.Exit(255)
	}

	// Set up the handler for the one and only endpoint, and start the HTTP
	// server.
	http.HandleFunc("/", qrHandler)

	fmt.Printf("Start listening on %s\n", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listening on %s: %s\n",
			listenAddr, err)
		os.Exit(255)
	}

	// HTTP servers never die :-)
}
