package ollama

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/stollenaar/ollamabot/internal/util"
)

func List() (ListResponse, error) {
	bodyData, err := do([]byte{}, "tags")

	if err != nil {
		return ListResponse{}, err
	}

	var r ListResponse
	json.Unmarshal([]byte(bodyData), &r)
	return r, nil
}

func Pull(pull PullRequest) error {
	data, err := json.Marshal(pull)
	if err != nil {
		return err
	}
	_, err = do(data, "pull")
	return err
}

func Generate(prompt GenerateRequest) (GenerateResponse, error) {

	data, err := json.Marshal(prompt)
	if err != nil {
		return GenerateResponse{}, err
	}

	bodyData, err := do(data, "generate")
	if err != nil {
		return GenerateResponse{}, err
	}
	var r GenerateResponse
	json.Unmarshal([]byte(bodyData), &r)
	return r, nil
}

func do(data []byte, endpoint string) (string, error) {
	// os.WriteFile("req.json", data, 0644)

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/%s", util.ConfigFile.OLLAMA_URL, endpoint), bytes.NewBuffer(data))

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	switch util.ConfigFile.OLLAMA_AUTH_TYPE {
	case "basic":
		username, err := util.GetOllamaUsername()
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		password, err := util.GetOllamaPassword()
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		token := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var bodyData string
	if resp != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		bodyData = buf.String()
	}
	if resp.StatusCode != 200 {
		fmt.Println(bodyData)
		return "", errors.New(bodyData)
	}
	return bodyData, nil
}
