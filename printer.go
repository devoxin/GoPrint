package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"net/url"

	"github.com/qeesung/image2ascii/convert"
)

var httpClient = http.Client{}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func render(options *convert.Options, image image.Image) string {
	width, height := 300.0, 200.0
	target := 2000.0
	divider := math.Sqrt(width*height) / math.Sqrt(target)
	newWidth, newHeight := int(math.Floor(width/divider)), int(math.Floor(height/divider))

	options.FixedHeight = newHeight
	options.FixedWidth = newWidth

	converter := convert.NewImageConverter()
	result := converter.Image2ASCIIString(image, options)

	return result
}

func downloadImage(url string) (image.Image, error) {
	res, err := httpClient.Get(url)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Server returned %d when fetching image", res.StatusCode)
	}

	defer res.Body.Close()

	imageData, imageType, err := image.Decode(res.Body)

	if err != nil {
		return nil, err
	}

	if imageType != "png" && imageType != "jpg" && imageType != "jpeg" {
		return nil, fmt.Errorf("Unsupported image type %s", imageType)
	}

	return imageData, nil
}

func handlePrinter(w http.ResponseWriter, req *http.Request) {
	keys, ok := req.URL.Query()["url"]

	if len(keys) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("You must provide a URL to an image"))
		return
	}

	imageURL := keys[0]

	if !ok || len(imageURL) == 0 || !isURL(imageURL) {
		w.WriteHeader(400)
		w.Write([]byte("Invalid URL provided"))
		return
	}

	image, err := downloadImage(imageURL)

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	convertOptions := convert.DefaultOptions
	convertOptions.Colored = false
	result := render(&convertOptions, image)

	w.WriteHeader(200)
	w.Write([]byte(result))
}

func main() {
	http.HandleFunc("/printer", handlePrinter)
	err := http.ListenAndServe(":7080", nil)

	if err != nil {
		panic(err)
	}
}
