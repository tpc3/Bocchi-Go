package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/tpc3/Bocchi-Go/lib/config"
)

const openai = "https://api.openai.com/v1/chat/completions"

var timeout *url.Error

func GptRequest(guild *config.Guild, msg *[]Message, num *int, data *config.Data) (string, error) {
	apikey := config.CurrentConfig.Chat.ChatToken
	response, err := getOpenAIResponse(guild, &apikey, msg, num, data)
	return response, err
}

func getOpenAIResponse(guild *config.Guild, apikey *string, messages *[]Message, num *int, data *config.Data) (string, error) {
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

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Duration(time.Duration(config.CurrentConfig.Guild.Timeout).Seconds()),
	}
	resp, err := client.Do(req)
	if err != nil {
		if errors.As(err, &timeout) && timeout.Timeout() {
			return "", err
		} else {
			log.Fatal("Sending http request error: ", err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 503 {
		var errorResponse ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			log.Panic("Decoding error response failed: ", err)
		}
		log.Print("Service is unavailable: ", errorResponse.Error.Message)
		return "", err
	}

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
	tokens := response.Usages.TotalTokens

	config.SaveData(data, tokens)

	return result, nil
}
