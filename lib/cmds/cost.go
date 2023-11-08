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
	Tokens_gpt_35_turbo_1106_prompt := float64(config.CurrentData.Tokens.Gpt_35_turbo_1106.Prompt)
	Tokens_gpt_35_turbo_1106_completion := float64(config.CurrentData.Tokens.Gpt_35_turbo_1106.Completion)
	Tokens_gpt_35_turbo_instruct_prompt := float64(config.CurrentData.Tokens.Gpt_35_turbo_instruct.Prompt)
	Tokens_gpt_35_turbo_instruct_completion := float64(config.CurrentData.Tokens.Gpt_35_turbo_instruct.Completion)
	Tokens_gpt_4_prompt := float64(config.CurrentData.Tokens.Gpt_4.Prompt)
	Tokens_gpt_4_cpmpletion := float64(config.CurrentData.Tokens.Gpt_4.Completion)
	Tokens_gpt_4_32k_prompt := float64(config.CurrentData.Tokens.Gpt_4_32k.Prompt)
	Tokens_gpt_4_32k_completion := float64(config.CurrentData.Tokens.Gpt_4_32k.Completion)
	Tokens_gpt_4_1106_preview_prompt := float64(config.CurrentData.Tokens.Gpt_4_1106_preview.Prompt)
	Tokens_gpt_4_1106_preview_completion := float64(config.CurrentData.Tokens.Gpt_4_1106_preview.Completion)
	Tokens_gpt_4_vision_preview_prompt := float64(config.CurrentData.Tokens.Gpt_4_vision_preview.Prompt)
	Tokens_gpt_4_vision_preview_completion := float64(config.CurrentData.Tokens.Gpt_4_vision_preview.Completion)
	Tokens_gpt_4_vision_preview_details := float64(config.CurrentData.Tokens.Gpt_4_vision_preview.Detail)

	if config.Lang[guild.Lang].Lang == "japanese" {
		rate = config.CurrentRate
	} else {
		rate = 1
	}

	cost_gpt_35_1106 := ((Tokens_gpt_35_turbo_1106_prompt / 1000) * 0.001 * rate) + ((Tokens_gpt_35_turbo_1106_completion / 1000) * 0.002 * rate)
	cost_gpt_35_instruct := ((Tokens_gpt_35_turbo_instruct_prompt / 1000) * 0.0015 * rate) + ((Tokens_gpt_35_turbo_instruct_completion / 1000) * 0.002 * rate)
	cost_gpt_4 := ((Tokens_gpt_4_prompt / 1000) * 0.03 * rate) + ((Tokens_gpt_4_cpmpletion / 1000) * 0.06 * rate)
	cost_gpt_4_32k := ((Tokens_gpt_4_32k_prompt / 1000) * 0.06 * rate) + ((Tokens_gpt_4_32k_completion / 1000) * 0.12 * rate)
	cost_gpt_4_1106_preview := ((Tokens_gpt_4_1106_preview_prompt / 1000) * 0.01 * rate) + ((Tokens_gpt_4_1106_preview_completion / 1000) * 0.03 * rate)
	cost_gpt_4_1106_vision_preview := ((Tokens_gpt_4_vision_preview_prompt / 1000) * 0.01 * rate) + ((Tokens_gpt_4_vision_preview_completion / 1000) * 0.03 * rate) + Tokens_gpt_4_vision_preview_details
	cost := (cost_gpt_35_1106 + cost_gpt_35_instruct + cost_gpt_4 + cost_gpt_4_32k + cost_gpt_4_1106_preview + cost_gpt_4_1106_vision_preview)
	coststr := strconv.FormatFloat(cost, 'f', 2, 64)
	if coststr == "0.00" {
		coststr = "0"
	}
	return coststr
}
