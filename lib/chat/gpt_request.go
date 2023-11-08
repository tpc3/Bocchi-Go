package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/tpc3/Bocchi-Go/lib/config"
)

const openai = "https://api.openai.com/v1/chat/completions"

var timeout *url.Error

func GptRequestImg(img *[]Img, data *config.Tokens, guild *config.Guild, topnum float64, tempnum float64, model string, detcost int) (string, error) {

	apikey := config.CurrentConfig.Chat.ChatToken

	requestBody := OpenaiRequestImg{
		Model:       model,
		Messages:    *img,
		Top_p:       topnum,
		Temperature: tempnum,
		MaxToken:    3000,
	}

	response, err := getOpenAIResponse(&apikey, data, model, requestBody, detcost)
	return response, err
}

func GptRequest(msg *[]Message, data *config.Tokens, guild *config.Guild, topnum float64, tempnum float64, model string, detail int) (string, error) {

	apikey := config.CurrentConfig.Chat.ChatToken

	requestBody := OpenaiRequest{
		Model:       model,
		Messages:    *msg,
		Top_p:       topnum,
		Temperature: tempnum,
	}

	response, err := getOpenAIResponse(&apikey, data, model, requestBody, detail)
	return response, err
}

func getOpenAIResponse(apikey *string, data *config.Tokens, model string, requestBody interface{}, detcost int) (string, error) {

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

	err = config.SaveData(data, &model, &promptTokens, &completionTokens, &totalTokens, &detcost)
	if err != nil {
		log.Fatal("Data save failed: ", err)
	}

	return result, nil
}
