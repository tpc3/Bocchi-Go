package config

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
)

type Data struct {
	Tokens Tokens
}

type Tokens struct {
	Gpt_35_turbo_1106     Gpt_35_turbo_1106
	Gpt_35_turbo_instruct Gpt_35_turbo_instruct
	Gpt_4                 Gpt_4
	Gpt_4_32k             Gpt_4_32k
	Gpt_4_1106_preview    Gpt_4_1106_preview
	Gpt_4_vision_preview  Gpt_4_vision_preview
}

type Gpt_35_turbo_1106 struct {
	Prompt     int
	Completion int
}

type Gpt_35_turbo_instruct struct {
	Prompt     int
	Completion int
}

type Gpt_4 struct {
	Prompt     int
	Completion int
}

type Gpt_4_32k struct {
	Prompt     int
	Completion int
}

type Gpt_4_1106_preview struct {
	Prompt     int
	Completion int
}

type Gpt_4_vision_preview struct {
	Prompt     int
	Completion int
	Detail     int
}

const dataFile = "./data.yml"

var (
	CurrentData Data
	CurrentRate float64
	mutex       sync.Mutex
)

func init() {
	GetRate()

	file, err := os.ReadFile(dataFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal("Data load failed: ", err)
		}

		initGpt_35_turbo_1106 := Gpt_35_turbo_1106{
			Prompt:     0,
			Completion: 0,
		}

		initGpt_35_turbo_instruct := Gpt_35_turbo_instruct{
			Prompt:     0,
			Completion: 0,
		}

		initGpt_4 := Gpt_4{
			Prompt:     0,
			Completion: 0,
		}

		initGpt_4_32k := Gpt_4_32k{
			Prompt:     0,
			Completion: 0,
		}

		initGpt_4_1106_preview := Gpt_4_1106_preview{
			Prompt:     0,
			Completion: 0,
		}

		initGpt_4_vision_preview := Gpt_4_vision_preview{
			Prompt:     0,
			Completion: 0,
			Detail:     0,
		}

		initTokens := Tokens{
			Gpt_35_turbo_1106:     initGpt_35_turbo_1106,
			Gpt_35_turbo_instruct: initGpt_35_turbo_instruct,
			Gpt_4:                 initGpt_4,
			Gpt_4_32k:             initGpt_4_32k,
			Gpt_4_1106_preview:    initGpt_4_1106_preview,
			Gpt_4_vision_preview:  initGpt_4_vision_preview,
		}

		CurrentData = Data{
			Tokens: initTokens,
		}

		initdata, err := yaml.Marshal(CurrentData)
		if err != nil {
			return
		}

		err = os.WriteFile(dataFile, initdata, 0666)
		if err != nil {
			return
		}
		return
	}

	err = yaml.Unmarshal(file, &CurrentData)
	if err != nil {
		log.Fatal("Data parse failed: ", err)
	}
}

func SaveData(data *Tokens, model *string, promptTokens *int, completionTokens *int, totalTokens *int, detcost *int) error {
	file, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &data)
	if err != nil {
		return err
	}

	switch *model {
	case "gpt-3.5-turbo":
		CurrentData.Tokens.Gpt_35_turbo_instruct.Prompt += *promptTokens
		CurrentData.Tokens.Gpt_35_turbo_instruct.Completion += *completionTokens
	case "gpt-3.5-turbo-instruct":
		CurrentData.Tokens.Gpt_35_turbo_instruct.Prompt += *promptTokens
		CurrentData.Tokens.Gpt_35_turbo_instruct.Completion += *completionTokens
	case "gpt-3.5-turbo-1106":
		CurrentData.Tokens.Gpt_35_turbo_1106.Prompt += *promptTokens
		CurrentData.Tokens.Gpt_35_turbo_1106.Completion += *completionTokens
	case "gpt-4":
		CurrentData.Tokens.Gpt_4.Prompt += *promptTokens
		CurrentData.Tokens.Gpt_4.Completion += *completionTokens
	case "gpt-4-32k":
		CurrentData.Tokens.Gpt_4_32k.Prompt += *promptTokens
		CurrentData.Tokens.Gpt_4_32k.Completion += *completionTokens
	case "gpt-4-1106-preview":
		CurrentData.Tokens.Gpt_4_1106_preview.Prompt += *promptTokens
		CurrentData.Tokens.Gpt_4_1106_preview.Completion += *completionTokens
	case "gpt-4-vision-preview":
		CurrentData.Tokens.Gpt_4_vision_preview.Prompt += *promptTokens
		CurrentData.Tokens.Gpt_4_vision_preview.Completion += *completionTokens
		CurrentData.Tokens.Gpt_4_vision_preview.Detail += *detcost
	}

	newData := Data{
		Tokens: CurrentData.Tokens,
	}

	mutex.Lock()
	writedata, err := yaml.Marshal(&newData)
	if err != nil {
		return nil
	}

	err = os.WriteFile(dataFile, writedata, 0666)
	if err != nil {
		return nil
	}
	mutex.Unlock()
	return nil
}

func RunCron() {
	c := cron.New()
	c.AddFunc("0 0 1 * *", func() { ResetTokens() })
	c.AddFunc("0 0 * * *", func() { GetRate() })
	c.Start()
}

func ResetTokens() error {
	CurrentData.Tokens.Gpt_35_turbo_1106.Prompt, CurrentData.Tokens.Gpt_35_turbo_1106.Completion, CurrentData.Tokens.Gpt_35_turbo_instruct.Prompt, CurrentData.Tokens.Gpt_35_turbo_instruct.Completion, CurrentData.Tokens.Gpt_4.Prompt, CurrentData.Tokens.Gpt_4.Completion, CurrentData.Tokens.Gpt_4_32k.Prompt, CurrentData.Tokens.Gpt_4_32k.Completion, CurrentData.Tokens.Gpt_4_1106_preview.Prompt, CurrentData.Tokens.Gpt_4_1106_preview.Completion, CurrentData.Tokens.Gpt_4_vision_preview.Prompt, CurrentData.Tokens.Gpt_4_vision_preview.Completion, CurrentData.Tokens.Gpt_4_vision_preview.Detail = 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0

	newData := Data{
		Tokens: CurrentData.Tokens,
	}

	mutex.Lock()
	data, err := yaml.Marshal(&newData)
	if err != nil {
		return nil
	}

	err = os.WriteFile(dataFile, data, 0666)
	if err != nil {
		return nil
	}
	mutex.Unlock()

	return nil
}

func GetRate() {
	url := "https://api.excelapi.org/currency/rate?pair=usd-jpy"
	resp, err := http.Get(url)

	if err != nil {
		log.Fatal("API for get rate error: ", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		CurrentRate = 130
		return
	}

	byteArray, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Reading body error: ", err)
	}

	CurrentRate, err = strconv.ParseFloat(string(byteArray), 64)
	if err != nil {
		log.Fatal("Parsing rate error: ", err)
	}
}
