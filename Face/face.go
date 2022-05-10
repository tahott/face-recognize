package face

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/Azure/go-autorest/autorest"
	"github.com/gofrs/uuid"
)

type FaceService struct {
	client                  *face.Client
	personGroupClient       *face.PersonGroupClient
	personGroupPersonClient *face.PersonGroupPersonClient
}

type GroupRegist struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type FaceRegist struct {
	Faces []string `json:"faces,omitempty"`
	Name  string   `json:"name,omitempty"`
}

func Regist(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	switch path[len(path)-1] {
	case "regist":
		switch r.Method {
		case "GET":
			var content []byte

			content, _ = ioutil.ReadFile("face_regist.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			w.Write(content)
		default:
			fmt.Fprint(w, "NOT ALLOWED METHOD")
		}
	case "group":
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

			fmt.Println(body.Name, body.Id)

			_, err = faceService.CreateGroup(body.Id, body.Name)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprintf(w, "Create group success!!")
		}
	default:
		fmt.Fprint(w, "NOT ALLOWED PATH")
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
		RecognitionModel: "recognition_04",
		Name:             &name,
	}

	return faceService.personGroupClient.Create(context.Background(), id, metadata)
}

func (faceService *FaceService) DeleteGroup(id string) {
	faceService.personGroupClient.Delete(context.Background(), id)
}

func (faceService *FaceService) CreatePerson(id, name, userData string) (face.Person, error) {
	metadata := face.NameAndUserDataContract{
		Name:     &name,
		UserData: &userData,
	}

	return faceService.personGroupPersonClient.Create(context.Background(), id, metadata)
}

func (faceService *FaceService) AddFaceData(group string, person uuid.UUID, url io.ReadCloser) {
	faceService.personGroupPersonClient.AddFaceFromStream(context.Background(), group, person, url, "", nil, "detection_03")
	faceService.personGroupClient.Train(context.Background(), group)
}
