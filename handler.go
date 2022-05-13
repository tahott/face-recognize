package main

import (
	"log"
	"net/http"
	"os"

	face "face-recognize/Face"
	addr "face-recognize/Page"
	public "face-recognize/Public"
)

func main() {
	listenAddr := ":8080"

	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}

	http.HandleFunc("/page/index", addr.Path)
	http.HandleFunc("/public/haarcascade_frontalface_default.xml", public.Classifier)

	http.HandleFunc("/face/regist", face.Regist)
	http.HandleFunc("/face/group", face.Group)
	http.HandleFunc("/face/identify", face.Identify)

	log.Printf("About to listen on %s. Go to https://127.0.0.1%s/", listenAddr, listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
