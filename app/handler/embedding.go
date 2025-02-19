package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	embeddingAPI        = "http://llm01-dev.ringover.net:8001/v1/embeddings"
	apiAuthHeader       = "Authorization"
	apiAuthToken        = "EMPTY"
	numWorkersEmbedding = 100
)

func GetEmbedding(input string) ([]float32, error) {
	log.Println("[INFO] GetEmbedding called with input: ", input)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "Yohhei/bge-multilingual-gemma2-gptq-4bit",
		"input": []string{input},
	})
	if err != nil {
		log.Println("[ERROR] Error while creating JSON payload: ", err)
		return nil, fmt.Errorf("error while creating JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", embeddingAPI, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("[ERROR] Error while creating API request: ", err)
		return nil, fmt.Errorf("error while creating API request: %w", err)
	}

	req.Header.Set(apiAuthHeader, apiAuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] Error while calling API: ", err)
		return nil, fmt.Errorf("error while calling API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("[ERROR] API returned status code: ", resp.StatusCode, " with body: ", string(body))
		return nil, fmt.Errorf("API returned status code %d with body %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println("[ERROR] Error while decoding JSON response: ", err)
		return nil, fmt.Errorf("error while decoding JSON response: %w", err)
	}

	if len(response.Data) == 0 {
		log.Println("[ERROR] No embeddings in the response")
		return nil, fmt.Errorf("no embeddings in the response")
	}

	return response.Data[0].Embedding, nil
}
