package face

import (
	b "bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func Identify(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		defer r.Body.Close()

		faceService := Start()

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		data := string(bytes)

		body := &FaceIdentify{}

		json.Unmarshal([]byte(data), body)

		i := strings.Index(body.Face, ",")
		dec, _ := base64.StdEncoding.DecodeString(body.Face[i+1:])
		reader := b.NewReader(dec)

		buf := make([]byte, len(dec))
		if _, err := reader.Read(buf); err != nil {
			log.Fatal(err)
		}

		r := io.NopCloser(strings.NewReader(string(buf)))

		personId, confidence := faceService.identifyFace(r)

		person, err := faceService.getPerson("face", personId)
		if err != nil {
			log.Fatal(err)
		}

		result := &ResponseIdentifyDto{
			Name:       *person.Name,
			Confidence: confidence,
		}

		json, _ := json.Marshal(result)

		w.Write(json)
	default:
		fmt.Fprint(w, "NOT ALLOWED METHOD")
	}
}

func Group(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		defer r.Body.Close()

		faceService := Start()

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		data := string(bytes)

		body := &GroupRegist{}

		json.Unmarshal([]byte(data), body)

		_, err = faceService.createGroup(body.Id, body.Name)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintf(w, "Create group success!!")
	case "DELETE":
		defer r.Body.Close()

		faceService := Start()

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		data := string(bytes)

		body := &GroupRegist{}

		json.Unmarshal([]byte(data), body)

		_, err = faceService.deleteGroup(body.Id)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintf(w, "Delete group")
	default:
		fmt.Fprint(w, "NOT ALLOWED METHOD")
	}
}

func Regist(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var content []byte

		content, _ = ioutil.ReadFile("face_regist.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		w.Write(content)
	case "POST":
		defer r.Body.Close()

		faceService := Start()

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		data := string(bytes)

		body := &FaceRegist{}

		json.Unmarshal([]byte(data), body)

		person, err := faceService.createPerson("face", body.Name, "")
		if err != nil {
			log.Fatal(err)
		}

		for _, image := range body.Faces {
			i := strings.Index(image, ",")
			dec, _ := base64.StdEncoding.DecodeString(image[i+1:])
			reader := b.NewReader(dec)

			buf := make([]byte, len(dec))

			if _, err := reader.Read(buf); err != nil {
				log.Fatal(err)
			}

			r := io.NopCloser(strings.NewReader(string(buf)))

			if _, err := faceService.addFaceData("face", *person.PersonID, r); err != nil {
				log.Fatal(err)
			}
		}

		faceService.personGroupClient.Train(context.Background(), "face")
	default:
		fmt.Fprint(w, "NOT ALLOWED METHOD")
	}
}
