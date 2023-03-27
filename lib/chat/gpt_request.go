package chat

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/tpc3/Bocchi-Go/lib/config"
)

const openai = "https://api.openai.com/v1/chat/completions"

func GptRequest(msg *string) string {
	apikey := config.CurrentConfig.Chat.ChatToken
	messages := []Message{
		{
			Role:    "user",
			Content: *msg,
		},
	}
	response := getOpenAIResponse(&apikey, &messages)
	return response
}

func getOpenAIResponse(apikey *string, messages *[]Message) string {
	requestBody := OpenaiRequest{
		Model:    "gpt-3.5-turbo",
		Messages: *messages,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal("Marshaling json error: ", err)
	}

	req, err := http.NewRequest("POST", openai, bytes.NewBuffer(requestJSON))
	if err != nil {
		log.Fatal("Creating http request error: ", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Sending http request error: ", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("Closing body error: ", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Reading body error: ", err)
	}

	var response OpenaiResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal("Unmarshaling json error: ", err)
	}

	result := response.Choices[0].Messages.Content

	return result
}
