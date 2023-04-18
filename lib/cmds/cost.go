package cmds

import (
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
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "🤔")
	embedMsg := embed.NewEmbed(session, orgMsg)
	embedMsg.Title = "Cost"
	embedMsg.Description = config.Lang[guild.Lang].Reply.Cost + calculationTokens(guild)
	session.MessageReactionRemove(orgMsg.ChannelID, orgMsg.ID, "🤔", session.State.User.ID)
	ReplyEmbed(session, orgMsg, embedMsg)
}

func calculationTokens(guild *config.Guild) string {

	var rate float64
	Tokens_gpt_35_turbo := float64(config.CurrentData.Tokens.Gpt_35_turbo)
	Tokens_gpt_4 := float64(config.CurrentData.Tokens.Gpt_4)
	Tokens_gpt_4_32k := float64(config.CurrentData.Tokens.Gpt_4_32k)

	if config.Lang[guild.Lang].Lang == "japanese" {
		rate = config.CurrentRate
	} else {
		rate = 1
	}

	cost_gpt_35 := (Tokens_gpt_35_turbo / 1000) * 0.0002 * rate
	cost_gpt_4 := (Tokens_gpt_4 / 1000) * 0.003 * rate
	cost_gpt_4_32k := (Tokens_gpt_4_32k / 1000) * 0.006 * rate
	cost := cost_gpt_35 + cost_gpt_4 + cost_gpt_4_32k
	coststr := strconv.FormatFloat(cost, 'f', 2, 64)
	if coststr == "0.00" {
		coststr = "0"
	}
	return coststr
}
