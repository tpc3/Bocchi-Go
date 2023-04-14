package cmds

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/embed"
)

const (
	Cost     = "cost"
	dataFile = "./data.yml"
)

func CostCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild) {
	embedMsg := embed.NewEmbed(session, orgMsg)
	embedMsg.Title = "Cost"
	embedMsg.Description = config.Lang[guild.Lang].Reply.Cost + calculationTokens(session, orgMsg, guild)
	ReplyEmbed(session, orgMsg, embedMsg)
}

func calculationTokens(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild) string {

	var rate float64
	Tokens := float64(config.CurrentData.Totaltokens)

	if config.Lang[guild.Lang].Lang == "japanese" {
		url := "https://api.excelapi.org/currency/rate?pair=usd-jpy"
		resp, err := http.Get(url)

		if err != nil {
			UnknownError(session, orgMsg, &guild.Lang, err)
		}

		defer resp.Body.Close()
		byteArray, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Reading body error: ", err)
		}

		rate, err = strconv.ParseFloat(string(byteArray), 64)
		if err != nil {
			log.Fatal("Parsing rate error: ", err)
		}
	} else {
		rate = 1
	}

	cost := (Tokens / 1000) * 0.002 * rate
	coststr := strconv.FormatFloat(cost, 'f', 2, 64)
	if coststr == "0.00" {
		coststr = "0"
	}
	return coststr
}
