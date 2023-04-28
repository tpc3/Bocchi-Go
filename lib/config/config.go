package config

import (
	"errors"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Debug   bool
	Discord struct {
		Token  string
		Status string
	}
	Chat          Chat
	Guild         Guild
	UserBlacklist []string `yaml:"user_blacklist"`
}

type Chat struct {
	ChatToken string
}

type Guild struct {
	Prefix  string
	Lang    string
	Model   string
	Timeout int
}

const configFile = "./config.yml"

var CurrentConfig Config
var CurrentLimitToken int

func init() {
	file, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("Config load failed: ", err)
	}
	err = yaml.Unmarshal(file, &CurrentConfig)
	if err != nil {
		log.Fatal("Config parse failed: ", err)
	}

	//verify
	if CurrentConfig.Debug {
		log.Print("Debug is enabled")
	}
	if CurrentConfig.Discord.Token == "" {
		log.Fatal("Token is empty")
	}

	loadLang()

	err = VerifyGuild(&CurrentConfig.Guild)
	if err != nil {
		log.Fatal("Config verify failed: ", err)
	}
}

func VerifyGuild(guild *Guild) error {
	if len(guild.Prefix) == 0 || len(guild.Prefix) >= 10 {
		return errors.New("prefix is too short or long")
	}
	_, exists := Lang[guild.Lang]
	if !exists {
		return errors.New("language does not exists")
	}
	return nil
}

func SaveGuild(guild *Guild) error {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	if guild.Prefix != CurrentConfig.Guild.Prefix || guild.Model != CurrentConfig.Guild.Model || guild.Lang != CurrentConfig.Guild.Lang || guild.Timeout != CurrentConfig.Guild.Timeout {
		CurrentConfig.Guild.Prefix = guild.Prefix
		CurrentConfig.Guild.Lang = guild.Lang
		CurrentConfig.Guild.Model = guild.Model
		CurrentConfig.Guild.Timeout = guild.Timeout
	}

	newConfig := Config{
		Debug:         config.Debug,
		Discord:       config.Discord,
		Chat:          config.Chat,
		Guild:         *guild,
		UserBlacklist: config.UserBlacklist,
	}
	data, err := yaml.Marshal(&newConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFile, data, 0666)
	if err != nil {
		return err
	}

	CurrentLimitToken = CheckLimitToken(guild.Model)

	return nil
}

func CheckLimitToken(model string) int {
	var token int

	switch model {
	case "gpt-3.5-turbo":
		token = 4096
	case "gpt-4":
		token = 8192
	case "gpt-4-32k":
		token = 32768
	}

	return token
}
