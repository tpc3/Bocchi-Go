package cmds

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/chat"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/embed"
)

const (
	Chat     = "chat"
	Continue = "-l "
)

var timeout *url.Error

func ChatCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, msg *string) {
	if *msg == "" {
		ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.SubCmd)
		return
	}
	var content, param string
	var num int = 0
	if strings.Contains(*msg, Continue) {
		if strings.HasPrefix(*msg, "-l ") {
			content = strings.SplitN(*msg, " ", 3)[2]
			param = strings.SplitN(*msg, " ", 3)[1]
			num, _ = strconv.Atoi(param)
		} else {
			content = strings.SplitN(*msg, " -l ", 2)[0]
			param = strings.SplitN(*msg, " -l ", 2)[1]
			num, _ = strconv.Atoi(param)
		}
	} else {
		content = *msg
	}
	start := time.Now()
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "🤔")
	response, coststr, err := chat.GptRequest(guild, &content, &num)
	if errors.As(err, &timeout) && timeout.Timeout() {
		ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.TimeOut)
		return
	}
	if utf8.RuneCountInString(response) > 4096 {
		ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.LongResponse)
		return
	}
	exec := time.Since(start).Seconds()
	dulation := strconv.FormatFloat(exec, 'f', 2, 64)
	embedMsg := embed.NewEmbed(session, orgMsg)
	if utf8.RuneCountInString(content) > 50 {
		embedMsg.Title = string([]rune(content)[:50]) + "..."
	} else {
		embedMsg.Title = content
	}
	embedMsg.Description = response
	exectimetext := config.Lang[guild.Lang].Reply.ExecTime
	second := config.Lang[guild.Lang].Reply.Second
	if config.CurrentConfig.Guild.ViewFees {
		embedMsg.Footer = &discordgo.MessageEmbedFooter{
			Text: exectimetext + dulation + second + "\n" + config.Lang[guild.Lang].Reply.Cost + coststr,
		}
	} else {
		embedMsg.Footer = &discordgo.MessageEmbedFooter{
			Text: exectimetext + dulation + second,
		}
	}
	session.MessageReactionRemove(orgMsg.ChannelID, orgMsg.ID, "🤔", session.State.User.ID)
	GPTReplyEmbed(session, orgMsg, embedMsg)
}
