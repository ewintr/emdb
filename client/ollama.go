package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Ollama struct {
	baseURL string
	c       *http.Client
}

func NewOllama(baseURL string) *Ollama {
	return &Ollama{
		baseURL: baseURL,
		c:       &http.Client{},
	}
}

func (o *Ollama) Generate(model, prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", o.baseURL)
	reqBody := struct {
		Model  string
		Prompt string
		Format string
		Stream bool
	}{
		Model:  model,
		Prompt: prompt,
		Format: "json",
		Stream: false,
	}
	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBodyJSON))
	if err != nil {
		return "", err
	}
	res, err := o.c.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	resBody := struct {
		Response string
	}{}
	if err := json.Unmarshal(body, &resBody); err != nil {
		return "", err
	}

	return resBody.Response, nil
}
