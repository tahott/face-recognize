package addr

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func Path(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	var content []byte

	switch path[len(path)-1] {
	case "index":
		content, _ = ioutil.ReadFile("index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	default:
	}

	w.Write(content)
}
