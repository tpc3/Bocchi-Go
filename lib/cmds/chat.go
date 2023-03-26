package cmds

import (
	"strconv"
	"time"
	"unicode/utf8"

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
	if utf8.RuneCountInString(response) > 1023 {
		ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.LongResponse)
	}
	exec := time.Since(start).Seconds()
	dulation := strconv.FormatFloat(exec, 'f', 2, 64)
	embedMsg := embed.NewEmbed(session, orgMsg)
	if utf8.RuneCountInString(msg) > 50 {
		embedMsg.Title = string([]rune(msg)[:50]) + "..."
	} else {
		embedMsg.Title = msg
	}
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
