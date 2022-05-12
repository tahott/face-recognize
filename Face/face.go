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
	Name string `json:"name,omitempty"`
}

type FaceRegist struct {
	Faces []string `json:"faces,omitempty"`
	Name  string   `json:"name,omitempty"`
}

type FaceIdentify struct {
	Face string `json:"face"`
}

type ResponseIdentifyDto struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

func Regist(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	switch path[len(path)-1] {
	case "identify":
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

			personId, confidence := faceService.IdentifyFace(r)

			person, err := faceService.GetPerson("face", personId)
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
	case "regist":
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

			person, err := faceService.CreatePerson("face", body.Name, "")
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

				if _, err := faceService.AddFaceData("face", *person.PersonID, r); err != nil {
					log.Fatal(err)
				}
			}

			faceService.personGroupClient.Train(context.Background(), "face")
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

			_, err = faceService.CreateGroup(body.Id, body.Name)
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

			_, err = faceService.DeleteGroup(body.Id)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprintf(w, "Delete group")
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

	personGroupPersonClient := face.NewPersonGroupPersonClient(os.Getenv("FACE_ENDPOINT"))
	personGroupPersonClient.Authorizer = authorizer

	faceService := &FaceService{
		client:                  &client,
		personGroupClient:       &personGroupClient,
		personGroupPersonClient: &personGroupPersonClient,
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

func (faceService *FaceService) DeleteGroup(id string) (autorest.Response, error) {
	return faceService.personGroupClient.Delete(context.Background(), id)
}

func (faceService *FaceService) CreatePerson(id, name, userData string) (face.Person, error) {
	metadata := face.NameAndUserDataContract{
		Name:     &name,
		UserData: &userData,
	}

	return faceService.personGroupPersonClient.Create(context.Background(), id, metadata)
}

func (faceService *FaceService) GetPerson(groupId string, personId uuid.UUID) (face.Person, error) {
	return faceService.personGroupPersonClient.Get(context.Background(), groupId, personId)
}

func (faceService *FaceService) AddFaceData(group string, person uuid.UUID, stream io.ReadCloser) (face.PersistedFace, error) {
	return faceService.personGroupPersonClient.AddFaceFromStream(context.Background(), group, person, stream, "", nil, "detection_03")
}

func (faceService *FaceService) IdentifyFace(stream io.ReadCloser) (uuid.UUID, float64) {
	returnIdentifyFaceID := true
	returnRecognitionModel := true
	returnFaceLandmarks := false

	detectedFaces, err := faceService.client.DetectWithStream(context.Background(), stream, &returnIdentifyFaceID, &returnFaceLandmarks, []face.AttributeType{"accessories"}, "recognition_04", &returnRecognitionModel, "detection_01")
	if err != nil {
		log.Fatal(err)
	}

	test, _ := json.Marshal(detectedFaces)
	fmt.Println(string(test))
	length := len(*detectedFaces.Value)
	testImageFaceIDs := make([]uuid.UUID, length)

	for i, f := range *detectedFaces.Value {
		testImageFaceIDs[i] = *f.FaceID
		res, _ := json.Marshal(f)
		fmt.Println(string(res))
	}

	personGroupID := "face"
	identifyRequestBody := face.IdentifyRequest{FaceIds: &testImageFaceIDs, PersonGroupID: &personGroupID}
	identifiedFaces, err := faceService.client.Identify(context.Background(), identifyRequestBody)
	if err != nil {
		log.Fatal(err)
	}

	var maxConfidence float64 = 0
	var target uuid.UUID
	iFaces := *identifiedFaces.Value
	for _, person := range iFaces {
		if len(*person.Candidates) > 0 {
			for _, candidate := range *person.Candidates {
				if *candidate.Confidence > maxConfidence {
					maxConfidence = *candidate.Confidence
					target = *candidate.PersonID
				}
			}
		}
	}

	return target, maxConfidence
}
