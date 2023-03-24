package config

import (
	"errors"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Debug   bool
	Help    string
	Discord struct {
		Token  string
		Status string
	}
	Chat          Chat
	Guild         Guild
	UserBlacklist []string `yaml:"user_blacklist"`
}

type Chat struct {
	ChatToken  string
	DeepLToken string
}

type Guild struct {
	Prefix      string
	Lang        string
	EnableDeepL bool
	MaxToken    int
}

const configFile = "./config.yml"

var CurrentConfig Config

func init() {
	loadLang()
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

func LoadGuild(id *string) (*Guild, error) {

	file, err := os.ReadFile(configFile)
	if os.IsNotExist(err) {
		return &Guild{
			Prefix: CurrentConfig.Guild.Prefix,
			Lang:   CurrentConfig.Guild.Lang,
		}, nil
	} else if err != nil {
		return nil, err
	}

	var guild Guild
	err = yaml.Unmarshal(file, &guild)
	if err != nil {
		return nil, err
	}

	return &guild, nil
}

func SaveGuild(id *string, guild *Guild) error {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	if guild.Prefix != CurrentConfig.Guild.Prefix || guild.Lang != CurrentConfig.Guild.Lang || guild.EnableDeepL != CurrentConfig.Guild.EnableDeepL || guild.MaxToken != CurrentConfig.Guild.MaxToken {
		CurrentConfig.Guild.Prefix = guild.Prefix
		CurrentConfig.Guild.Lang = guild.Lang
		CurrentConfig.Guild.EnableDeepL = guild.EnableDeepL
		CurrentConfig.Guild.MaxToken = guild.MaxToken
	}

	newConfig := Config{
		Debug:         config.Debug,
		Help:          config.Help,
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
	return nil
}
