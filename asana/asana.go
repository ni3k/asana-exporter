package asanaexporter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type AsanaEntity string

const (
	UsersEntity    AsanaEntity = "users"
	ProjectsEntity AsanaEntity = "projects"
)

var ErrFailedRetry = errors.New("missing retry after header on limited request")

type AsanaExtractor struct {
	baseUrl  string
	apiToken string
	client   *http.Client
}

func NewAsanaExtractor(baseUrl string, apiToken string) *AsanaExtractor {
	client := http.Client{}
	return &AsanaExtractor{baseUrl: baseUrl, apiToken: apiToken, client: &client}
}

type EntityFilters struct {
	Workspace string
	Team      string
	Limit     int
	Offset    int
}

type AsanaResponse struct {
	Data   []ObjectSchema `json:"data"`
	Errors []ErrorSchema  `json:"errors"`
}

type ErrorSchema struct {
	Message string `json:"message"`
}

type ObjectSchema struct {
	Gid          string `json:"gid"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
}

func (a *AsanaExtractor) RetrieveEntities(entity string) (*AsanaResponse, error) {
	entityUrl, err := a.newAsanaUrl(entity)
	if err != nil {
		return nil, err
	}
	// todo: implement limit & offset along with worskapce id

	req, err := a.newRequest(entityUrl)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 429 {
		retryAfter, err := strconv.Atoi(resp.Header.Get("retry-after"))
		if err != nil || retryAfter == 0 {
			return nil, ErrFailedRetry
		}
		retryAfterTimer := time.NewTimer(time.Second * time.Duration(retryAfter))
		<-retryAfterTimer.C
		return a.RetrieveEntities(entity)
	}

	entitiesResponse := new(AsanaResponse)
	if err = json.NewDecoder(resp.Body).Decode(entitiesResponse); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return entitiesResponse, nil
}

func (a *AsanaExtractor) StoreEntities(entity string, entities *AsanaResponse) error {
	filePath := fmt.Sprintf("exports/%s_%d.json", entity, time.Now().UTC().Unix())
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(entities.Data)
	return err
}

func (a *AsanaExtractor) newAsanaUrl(path string) (*url.URL, error) {
	path, err := url.JoinPath(a.baseUrl, path)
	if err != nil {
		return nil, err
	}

	asanaUrl, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	return asanaUrl, nil
}

func (a *AsanaExtractor) newRequest(u *url.URL) (*http.Request, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+a.apiToken)
	req.Header.Set("accept", "application/json")

	return req, nil
}
