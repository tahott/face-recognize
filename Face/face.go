package face

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type FaceRegist struct {
	Faces []string `json:"faces,omitempty"`
	Name  string   `json:"name,omitempty"`
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

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		data := string(bytes)

		body := &FaceRegist{}

		json.Unmarshal([]byte(data), body)

		fmt.Println(body.Name, len(body.Faces))
	default:
		fmt.Fprint(w, "NOT ALLOWED METHOD")
	}
}
