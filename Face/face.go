package face

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/Azure/go-autorest/autorest"
)

type FaceService struct {
	client                  *face.Client
	personGroupClient       *face.PersonGroupClient
	personGroupPersonClient *face.PersonGroupPersonClient
}

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

func Start() *FaceService {
	authorizer := autorest.NewCognitiveServicesAuthorizer(os.Getenv("FACE_SUB_KEY"))

	client := face.NewClient(os.Getenv("FACE_ENDPOINT"))
	client.Authorizer = authorizer

	personGroupClient := face.NewPersonGroupClient(os.Getenv("FACE_ENDPOINT"))
	personGroupClient.Authorizer = authorizer

	faceService := &FaceService{
		client:            &client,
		personGroupClient: &personGroupClient,
	}

	return faceService
}

func (faceService *FaceService) CreateGroup(id, name string) (autorest.Response, error) {
	metadata := face.MetaDataContract{
		RecognitionModel: "recongnition_04",
		Name:             &name,
	}

	return faceService.personGroupClient.Create(context.Background(), id, metadata)
}

func (faceService *FaceService) CreatePerson(id, name, userData string) (face.Person, error) {
	metadata := face.NameAndUserDataContract{
		Name:     &name,
		UserData: &userData,
	}

	return faceService.personGroupPersonClient.Create(context.Background(), id, metadata)
}
