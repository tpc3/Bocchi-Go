package config

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
)

type Data struct {
	Totaltokens int
}

const dataFile = "./data.yml"

var CurrentData Data

var mutex sync.Mutex

func init() {
	file, err := os.ReadFile(dataFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal("Data load failed: ", err)
		}
		return
	}
	err = yaml.Unmarshal(file, &CurrentData)
	if err != nil {
		log.Fatal("Data parse failed: ", err)
	}
}

func SaveData(data *Data, tokens int) error {
	file, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &data)
	if err != nil {
		return err
	}

	if CurrentData.Totaltokens != tokens {
		CurrentData.Totaltokens = CurrentData.Totaltokens + tokens
	}

	newData := Data{
		Totaltokens: CurrentData.Totaltokens,
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
	c.Start()
}

func ResetTokens() error {
	CurrentData.Totaltokens = 0

	newData := Data{
		Totaltokens: CurrentData.Totaltokens,
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
