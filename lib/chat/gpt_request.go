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
	"strconv"
	"time"

	"github.com/tpc3/Bocchi-Go/lib/config"
)

const openai = "https://api.openai.com/v1/chat/completions"

var timeout *url.Error

func GptRequest(guild *config.Guild, msg *[]Message, num *int) (response string, coststr string, err error) {
	apikey := config.CurrentConfig.Chat.ChatToken
	response, coststr, err = getOpenAIResponse(guild, &apikey, msg, num)
	return
}

func getOpenAIResponse(guild *config.Guild, apikey *string, messages *[]Message, num *int) (string, string, error) {
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
			return "", "", err
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
		return "", "", err
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
	cost := calculationCost(tokens, guild)

	return result, cost, nil
}

func calculationCost(tokens int, guild *config.Guild) string {
	rate := getRate(guild)
	cost := (float64(tokens) / 1000) * 0.002 * rate
	recost := strconv.FormatFloat(cost, 'f', 2, 64)
	if recost == "0.00" {
		i := 1
		text := "0.00"
		for {
			recost = strconv.FormatFloat(cost, 'f', 2+i, 64)
			text = text + "0"
			if recost != text {
				break
			}
		}
	}
	return recost
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
