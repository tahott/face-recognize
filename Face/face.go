package face

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

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

func (faceService *FaceService) createGroup(id, name string) (autorest.Response, error) {
	metadata := face.MetaDataContract{
		RecognitionModel: "recognition_04",
		Name:             &name,
	}

	return faceService.personGroupClient.Create(context.Background(), id, metadata)
}

func (faceService *FaceService) deleteGroup(id string) (autorest.Response, error) {
	return faceService.personGroupClient.Delete(context.Background(), id)
}

func (faceService *FaceService) createPerson(id, name, userData string) (face.Person, error) {
	metadata := face.NameAndUserDataContract{
		Name:     &name,
		UserData: &userData,
	}

	return faceService.personGroupPersonClient.Create(context.Background(), id, metadata)
}

func (faceService *FaceService) getPerson(groupId string, personId uuid.UUID) (face.Person, error) {
	return faceService.personGroupPersonClient.Get(context.Background(), groupId, personId)
}

func (faceService *FaceService) addFaceData(group string, person uuid.UUID, stream io.ReadCloser) (face.PersistedFace, error) {
	return faceService.personGroupPersonClient.AddFaceFromStream(context.Background(), group, person, stream, "", nil, "detection_03")
}

func (faceService *FaceService) identifyFace(stream io.ReadCloser) (uuid.UUID, float64) {
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
