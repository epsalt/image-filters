package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/epsalt/image-filters/image"
)

type Base64Image struct {
	ImageData string `json:"data"`
}

func greyscale(w http.ResponseWriter, req *http.Request) {
	var d Base64Image

	if req.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&d)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m, err := image.Decode(d.ImageData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m.Greyscale()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Base64Image{image.Encode(m)})
}

func main() {
	http.HandleFunc("/greyscale", greyscale)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
