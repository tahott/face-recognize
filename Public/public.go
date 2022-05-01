package public

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func Classifier(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	var content []byte

	switch path[len(path)-1] {
	case "haarcascade_frontalface_default.xml":
		content, _ = ioutil.ReadFile("haarcascade_frontalface_default.xml")
		w.Header().Set("Content-Type", "application/xml;")
	default:
	}

	w.Write(content)
}
