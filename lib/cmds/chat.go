package cmds

import (
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/chat"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/embed"
	"github.com/tpc3/Bocchi-Go/lib/utils"
)

const (
	Chat     = "chat"
	Continue = "-l "
)

var timeout *url.Error

func ChatCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, msg *string, data *config.Data) {
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

	msgChain := []chat.Message{{Role: "user", Content: *msg}}

	if orgMsg.ReferencedMessage != nil {
		loopTargetMsg, err := session.ChannelMessage(orgMsg.ChannelID, orgMsg.ReferencedMessage.ID)
		if err != nil {
			UnknownError(session, orgMsg, &guild.Lang, err)
			return
		}
		// Get reply msgs recursively
		for i := 0; i < num; i++ {
			if loopTargetMsg.Author.ID != session.State.User.ID {
				break
			} else if loopTargetMsg.Embeds[0].Color != embed.ColorGPT { //ToDo: Better handling
				break
			}
			msgChain = append(msgChain, chat.Message{Role: "assistant", Content: loopTargetMsg.Embeds[0].Description})

			if loopTargetMsg.ReferencedMessage == nil {
				break
			}
			loopTargetMsg, err = session.ChannelMessage(orgMsg.ChannelID, loopTargetMsg.ReferencedMessage.ID)
			if err != nil {
				UnknownError(session, orgMsg, &guild.Lang, err)
				return
			}
			if loopTargetMsg.Author.ID != orgMsg.Author.ID {
				break
			} else if loopTargetMsg.Content == "" {
				break
			}
			_, trimmed := utils.TrimPrefix(loopTargetMsg.Content, config.CurrentConfig.Guild.Prefix, session.State.User.Mention())
			msgChain = append(msgChain, chat.Message{Role: "user", Content: trimmed})

			if loopTargetMsg.ReferencedMessage == nil {
				break
			}
			loopTargetMsg, err = session.ChannelMessage(orgMsg.ChannelID, loopTargetMsg.ReferencedMessage.ID)
			if err != nil {
				UnknownError(session, orgMsg, &guild.Lang, err)
				return
			}
		}

		reverse(msgChain)
	}

	start := time.Now()
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "🤔")
	response, err := chat.GptRequest(guild, &msgChain, &num, data)
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
	embedMsg.Footer = &discordgo.MessageEmbedFooter{
		Text: exectimetext + dulation + second,
	}
	session.MessageReactionRemove(orgMsg.ChannelID, orgMsg.ID, "🤔", session.State.User.ID)
	GPTReplyEmbed(session, orgMsg, embedMsg)
}

// https://stackoverflow.com/questions/28058278/how-do-i-reverse-a-slice-in-go
func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
