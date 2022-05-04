package face

import (
	"io/ioutil"
	"net/http"
)

func Regist(w http.ResponseWriter, r *http.Request) {
	var content []byte

	content, _ = ioutil.ReadFile("face_regist.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.Write(content)
}
