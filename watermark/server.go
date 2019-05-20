package main

import (
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"log"
	"net/http"
)

func main() {
	registerRoutes()

	fmt.Println("server is listening on http://localhost:3210")
	err := http.ListenAndServe(":3210", logRequest(http.DefaultServeMux))
	if err != nil {
		panic(err)
	}
}

func registerRoutes() {
	http.HandleFunc("/", info)
	http.HandleFunc("/watermark", watermark)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println(err)
	fmt.Fprintf(w, err.Error())
}

func info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "send POST request to /watermark")
	fmt.Fprintln(w, "image - the main image file that will be marked")
	fmt.Fprintln(w, "watermark - the watermark file")
}

func watermark(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		info(w, r)
		return
	}

	r.ParseMultipartForm(32 << 20)

	imageDecoded, _, err := getImageDecoded(r, "image")
	if err != nil {
		handleError(w, err)
		return
	}

	width, height := getResizeWidthAndHeight(imageDecoded)
	resize := imaging.Resize(imageDecoded, width, height, imaging.Lanczos)

	watermarkDecoded, _, err := getImageDecoded(r, "watermark")
	if err != nil {
		handleError(w, err)
		return
	}

	resizeBounds := resize.Bounds()
	resizeSize := resizeBounds.Size()
	watermarkBounds := watermarkDecoded.Bounds()
	watermarkSize := watermarkBounds.Size()

	offsetX := (resizeSize.X - watermarkSize.X) / 2
	offsetY := (resizeSize.Y - watermarkSize.Y) / 2

	relX := resizeSize.X / watermarkSize.X
	relY := resizeSize.Y / watermarkSize.Y

	mX := watermarkSize.X + (watermarkSize.X/relX)/2 + 15
	mY := watermarkSize.Y + (watermarkSize.Y/relY)/2 + 15

	draw.Draw(resize, resizeBounds, resize, image.ZP, draw.Src)
	draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetX, offsetY)), watermarkDecoded, image.ZP, draw.Over)

	/* mark the central meridian */
	offsetYU := offsetY
	offsetYD := offsetY
	for i := relY; i > 0; i-- {
		offsetYU = offsetYU + mY
		offsetYD = offsetYD - mY
		draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetX, offsetYU)), watermarkDecoded, image.ZP, draw.Over)
		draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetX, offsetYD)), watermarkDecoded, image.ZP, draw.Over)
	}

	offsetXA := offsetX
	offsetXB := offsetX
	for i := relX; i > 0; i-- {
		offsetXA = offsetXA + mX
		draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetXA, offsetY)), watermarkDecoded, image.ZP, draw.Over)

		offsetXB = offsetXB - mX
		draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetXB, offsetY)), watermarkDecoded, image.ZP, draw.Over)

		offsetYA := offsetY
		offsetYB := offsetY
		for i := relY; i > 0; i-- {
			offsetYA = offsetYA + mY
			draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetXA, offsetYA)), watermarkDecoded, image.ZP, draw.Over)
			draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetXB, offsetYA)), watermarkDecoded, image.ZP, draw.Over)

			offsetYB = offsetYB - mY
			draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetXA, offsetYB)), watermarkDecoded, image.ZP, draw.Over)
			draw.Draw(resize, watermarkBounds.Add(image.Pt(offsetXB, offsetYB)), watermarkDecoded, image.ZP, draw.Over)
		}
	}

	err = jpeg.Encode(w, resize, &jpeg.Options{Quality: 95})
	if err != nil {
		handleError(w, err)
	}
}

func getImageDecoded(r *http.Request, name string) (decoded image.Image, ext string, err error) {
	imageFile, _, err := r.FormFile(name)
	if err != nil {
		return decoded, ext, err
	}

	err = imageFile.Close()
	if err != nil {
		return decoded, ext, err
	}

	decoded, ext, err = image.Decode(imageFile)
	if err != nil {
		return decoded, ext, err
	}

	return decoded, ext, nil
}

func getResizeWidthAndHeight(img image.Image) (width int, height int) {
	imageSize := img.Bounds().Size()

	if imageSize.X > imageSize.Y {
		width = 1024
	} else {
		height = 768
	}

	return width, height
}
