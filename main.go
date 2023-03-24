package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/handler"
)

func main() {
	Token := config.CurrentConfig.Discord.Token
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("error creating Discord session: ", err)
	}
	discord.AddHandler(handler.MessageCreate)
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent)
	err = discord.Open()
	if err != nil {
		log.Fatal("error opening connection: ", err)
	}
	discord.UpdateGameStatus(0, config.CurrentConfig.Discord.Status)
	log.Print("Bocchi-Go is now running!")
	defer discord.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Print("Bocchi-Go is gracefully shutdowning!")
}
