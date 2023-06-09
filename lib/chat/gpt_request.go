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

func GptRequest(msg *[]Message, data *config.Tokens, guild *config.Guild, topnum float64, tempnum float64, model string) (string, error) {
	apikey := config.CurrentConfig.Chat.ChatToken
	response, err := getOpenAIResponse(&apikey, msg, data, guild, topnum, tempnum, model)
	return response, err
}

func getOpenAIResponse(apikey *string, messages *[]Message, data *config.Tokens, guild *config.Guild, topnum float64, tempnum float64, model string) (string, error) {
	requestBody := OpenaiRequest{
		Model:       model,
		Messages:    *messages,
		Top_p:       topnum,
		Temperature: tempnum,
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
				Timeout:   time.Duration(config.CurrentConfig.Guild.Timeout) * time.Second,
				KeepAlive: time.Duration(config.CurrentConfig.Guild.Timeout) * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   time.Duration(config.CurrentConfig.Guild.Timeout) * time.Second,
			ResponseHeaderTimeout: time.Duration(config.CurrentConfig.Guild.Timeout) * time.Second,
			ExpectContinueTimeout: time.Duration(config.CurrentConfig.Guild.Timeout) * time.Second,
		},
		Timeout: time.Duration(config.CurrentConfig.Guild.Timeout) * time.Second,
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
	if resp.StatusCode != 200 {
		var errorResponse ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			log.Panic("Decoding error response failed: ", err)
		}
		log.Print("API error: ", errorResponse.Error.Message)
		err = errors.New(errorResponse.Error.Message)
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
	promptTokens := response.Usages.PromptTokens
	completionTokens := response.Usages.CompletionTokens
	totalTokens := response.Usages.TotalTokens

	err = config.SaveData(data, &model, &promptTokens, &completionTokens, &totalTokens)
	if err != nil {
		log.Fatal("Data save failed: ", err)
	}

	return result, nil
}
