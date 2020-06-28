package main

import (
    "errors"
    "fmt"
    "github.com/qeesung/image2ascii/convert"
    "image"
    "net/http"
    "net/url"
    _ "image/jpeg"
    _ "image/png"
)

func isURL(str string) bool {
    u, err := url.Parse(str)
    return err == nil && u.Scheme != "" && u.Host != ""
}

func render(options *convert.Options, image image.Image) (string, error) {
    converter := convert.NewImageConverter()
    result := converter.Image2ASCIIString(image, options)

    if len(result) > 2000 {
        if options.Ratio <= 0.0 {
            return "", errors.New("Couldn't scale image below 2000 characters")
        }

        options.Ratio -= 0.05
        return render(options, image)
    }
    
    return result, nil
}

func downloadImage(url string) (image.Image, error) {
    res, err := http.Get(url)

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

    if imageType != "png" && imageType != "jpg" {
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
    result, err := render(&convertOptions, image)

    if err != nil {
        w.WriteHeader(400)
        w.Write([]byte(err.Error()))
        return
    }

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
