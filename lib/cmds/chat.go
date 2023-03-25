package cmds

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/chat"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/embed"
)

const (
	Chat      = "chat"
	parameter = "-t "
)

func ChatCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, param *string) {
	msg := *param
	start := time.Now()
	response := chat.GptRequest(&msg)
	exec := time.Since(start).Seconds()
	dulation := strconv.FormatFloat(exec, 'f', 2, 64)
	embedMsg := embed.NewEmbed(session, orgMsg)
	embedMsg.Title = msg
	embedMsg.Fields = append(embedMsg.Fields, &discordgo.MessageEmbedField{
		Value: response,
	})
	exectimetext := config.Lang[guild.Lang].Reply.ExecTime
	second := config.Lang[guild.Lang].Reply.Second
	embedMsg.Footer = &discordgo.MessageEmbedFooter{
		Text: exectimetext + dulation + second,
	}
	GPTReplyEmbed(session, orgMsg, embedMsg)
}
