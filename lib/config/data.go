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
	Gpt_35_turbo int
	Gpt_4        int
	Gpt_4_32k    int
}

const dataFile = "./data.yml"

var (
	CurrentData Data
	CurrentRate float64
	mutex       sync.Mutex
)

func init() {
	RunCron()
	GetRate()

	file, err := os.ReadFile(dataFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal("Data load failed: ", err)
		}

		initTokens := Tokens{
			Gpt_35_turbo: 0,
			Gpt_4:        0,
			Gpt_4_32k:    0,
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

func SaveData(data *Tokens, model *string, tokens *int) error {
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
		CurrentData.Tokens.Gpt_35_turbo += *tokens
	case "gpt-4":
		CurrentData.Tokens.Gpt_4 += *tokens
	case "gpt-4-32k":
		CurrentData.Tokens.Gpt_4_32k += *tokens
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
	CurrentData.Tokens.Gpt_35_turbo, CurrentData.Tokens.Gpt_4, CurrentData.Tokens.Gpt_4_32k = 0, 0, 0

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
