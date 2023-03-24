package cmds

import (
	"strconv"
	"strings"
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
	var deepl, content, enabledeepltext string
	if strings.Contains(msg, parameter) {
		if strings.HasPrefix(msg, "-t ") {
			deepl = strings.SplitN(msg, parameter, -1)[1]
			deepl = strings.SplitN(deepl, " ", -1)[0]
			content = strings.SplitN(msg, " ", -1)[2]
		} else {
			deepl = strings.SplitN(msg, parameter, -1)[1]
			deepl = strings.SplitN(deepl, " ", -1)[0]
			content = strings.SplitN(msg, " -t ", -1)[0]
		}
	} else {
		content = msg
	}
	start := time.Now()
	response := chat.GptRequest(session, orgMsg, &content)
	exec := time.Since(start).Seconds()
	dulation := strconv.FormatFloat(exec, 'f', 2, 64)
	if deepl == "true" {
		chat.Deepl(&msg)
		enabledeepltext = config.Lang[guild.Lang].Reply.DeeplEnable
	} else {
		enabledeepltext = config.Lang[guild.Lang].Reply.DeeplDisable
	}
	embedMsg := embed.NewEmbed(session, orgMsg)
	embedMsg.Title = content
	embedMsg.Fields = append(embedMsg.Fields, &discordgo.MessageEmbedField{
		Value: response,
	})
	exectimetext := config.Lang[guild.Lang].Reply.ExecTime
	second := config.Lang[guild.Lang].Reply.Second
	enabledeepltitle := config.Lang[guild.Lang].Reply.Deepl
	embedMsg.Footer = &discordgo.MessageEmbedFooter{
		Text: exectimetext + dulation + second + "\n" + enabledeepltitle + enabledeepltext,
	}
	GPTReplyEmbed(session, orgMsg, embedMsg)
}
