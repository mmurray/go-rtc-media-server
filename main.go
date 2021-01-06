package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"regexp"

	"github.com/pion/webrtc/v3"
	"github.com/rs/xid"
)

var (
	validParam  = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	outputDir   = flag.String("o", "/tmp", "Output directory")
	addr        = flag.String("addr", ":8888", "Bind address for web server")
	maxParamLen = flag.Int("max_param_len", 64, "Max fname URL parameter length")
)

func offerHandler(w http.ResponseWriter, r *http.Request) {
	// Allow CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// Decode the JSON body
	offer := webrtc.SessionDescription{}
	err := json.NewDecoder(r.Body).Decode(&offer)
	if err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create the RTC connection
	filename := xid.New().String() + ".webm"
	filenameParam := r.URL.Query().Get("fname")
	if filenameParam != "" && len(filenameParam) < *maxParamLen && validParam.Match([]byte(filenameParam)) {
		filename = filenameParam + "_" + filename
	}
	saver := newWebmSaver(path.Join(*outputDir, filename))
	peerConnection := createWebRTCConn(saver, offer)

	// Marshal the JSON response and write it
	ld, err := json.Marshal(peerConnection.LocalDescription())
	if err != nil {
		http.Error(w, "Error unmarshaling: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(ld)
}

func main() {
	flag.Parse()
	http.HandleFunc("/offer", offerHandler)
	fmt.Println("Starting web server on " + *addr)
	fmt.Println("(Outputting videos to: " + *outputDir + ")")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
