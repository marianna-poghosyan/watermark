package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	imageFlag := flag.String("image", "image.jpg", "image filename")
	watermarkFlag := flag.String("watermark", "watermark.png", "watermark filename")
	flag.Parse()

	MakeRequest(*imageFlag, *watermarkFlag)
}

func MakeRequest(img string, watermark string) {
	imgFile := getFile(img)
	defer imgFile.Close()

	watermarkFile := getFile(watermark)
	defer watermarkFile.Close()

	var body bytes.Buffer

	multiPartWriter := multipart.NewWriter(&body)

	imageWriter, err := multiPartWriter.CreateFormFile("image", img)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = io.Copy(imageWriter, imgFile)
	if err != nil {
		log.Fatalln(err)
	}

	wmWriter, err := multiPartWriter.CreateFormFile("watermark", watermark)
	_, err = io.Copy(wmWriter, watermarkFile)
	if err != nil {
		log.Fatalln(err)
	}

	multiPartWriter.Close()

	req := getRequest(&body)
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer response.Body.Close()

	file, err := os.Create("marked.jpg")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer file.Close()

	// Write the body to file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func getFile(name string) (file *os.File) {
	file, err := os.Open(name)
	if err != nil {
		log.Fatalln(err)
	}

	return file
}

func getRequest(body io.Reader) (req *http.Request) {
	req, err := http.NewRequest("POST", "http://localhost:3210/watermark", body)
	if err != nil {
		log.Fatalln(err)
	}

	return req
}