package main

import (
	"encoding/json"
	"image/color/palette"
	"log"
	"net/http"

	"github.com/epsalt/image-filters/image"
)

type Base64Image struct {
	ImageData string `json:"data"`
}

func filterHandler(filter func(image.RGBA)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
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

		filter(m)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Base64Image{image.Encode(m)})
	}
}

func main() {
	http.HandleFunc("/greyscale", filterHandler(func(m image.RGBA) { m.Greyscale() }))
	http.HandleFunc("/blur", filterHandler(func(m image.RGBA) { m.Blur(0.1, 3) }))
	http.HandleFunc("/dither", filterHandler(func(m image.RGBA) { m.Dither(palette.Plan9) }))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
