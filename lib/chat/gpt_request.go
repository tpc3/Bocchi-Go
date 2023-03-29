package chat

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/tpc3/Bocchi-Go/lib/config"
)

const openai = "https://api.openai.com/v1/chat/completions"

func GptRequest(guild *config.Guild, msg *string) (string, string) {
	apikey := config.CurrentConfig.Chat.ChatToken
	messages := []Message{
		{
			Role:    "user",
			Content: *msg,
		},
	}
	response, coststr := getOpenAIResponse(guild, &apikey, &messages)
	return response, coststr
}

func getOpenAIResponse(guild *config.Guild, apikey *string, messages *[]Message) (string, string) {
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
	tokens := response.Usages.TotalTokens
	cost := calculationCost(tokens, guild)

	return result, cost
}

func calculationCost(tokens int, guild *config.Guild) string {
	rate := getRate(guild)
	cost := (float64(tokens) / 1000) * 0.002 * rate
	return strconv.FormatFloat(cost, 'f', 2, 64)
}

func getRate(guild *config.Guild) float64 {
	if config.Lang[guild.Lang].Lang == "japanese" {
		url := "https://api.excelapi.org/currency/rate?pair=usd-jpy"
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal("Sending http request error: ", err)
		}
		defer resp.Body.Close()
		byteArray, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Reading body error: ", err)
		}
		rate, err := strconv.ParseFloat(string(byteArray), 64)
		if err != nil {
			log.Fatal("Parsing rate error: ", err)
		}
		return rate
	} else {
		return 1
	}
}
