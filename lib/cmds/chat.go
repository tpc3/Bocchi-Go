package cmds

import (
	"errors"
	"log"
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

func ChatCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, msg *string, data *config.Tokens) {
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "🤔")
	if *msg == "" {
		ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.SubCmd)
		return
	}
	var content, param string
	var num int
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
		if strings.HasPrefix(content, "-l ") {
			content = strings.TrimPrefix(content, "-l ")
		} else {
			content = strings.TrimSuffix(content, " -l")
		}
		num = 2
	}

	msgChain := []chat.Message{{Role: "user", Content: content}}

	if orgMsg.ReferencedMessage != nil {
		loopTargetMsg, err := session.State.Message(orgMsg.ChannelID, orgMsg.ReferencedMessage.ID)
		if err != nil {
			loopTargetMsg, err = session.ChannelMessage(orgMsg.ChannelID, orgMsg.ReferencedMessage.ID)
			if err != nil {
				log.Panic("Failed to get channel message: ", err)
			}
			err = session.State.MessageAdd(loopTargetMsg)
			if err != nil {
				log.Panic("Failed to add message into state: ", err)
			}
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

			PrevState := loopTargetMsg
			loopTargetMsg, err = session.State.Message(orgMsg.ChannelID, loopTargetMsg.ReferencedMessage.ID)
			if err != nil {
				loopTargetMsg, err = session.ChannelMessage(orgMsg.ChannelID, PrevState.ReferencedMessage.ID)
				if err != nil {
					log.Panic("Failed to get channel message: ", err)
				}
				err = session.State.MessageAdd(loopTargetMsg)
				if err != nil {
					log.Panic("Failed to add message into state: ", err)
				}
			}

			if loopTargetMsg.Author.ID != orgMsg.Author.ID {
				break
			} else if loopTargetMsg.Content == "" {
				break
			}
			_, trimmed := utils.TrimPrefix(loopTargetMsg.Content, config.CurrentConfig.Guild.Prefix+Chat+" ", session.State.User.Mention())
			msgChain = append(msgChain, chat.Message{Role: "user", Content: trimmed})

			if loopTargetMsg.ReferencedMessage == nil {
				break
			}

			PrevState = loopTargetMsg
			loopTargetMsg, err = session.State.Message(orgMsg.ChannelID, loopTargetMsg.ReferencedMessage.ID)
			if err != nil {
				loopTargetMsg, err = session.ChannelMessage(orgMsg.ChannelID, PrevState.ReferencedMessage.ID)
				if err != nil {
					log.Panic("Failed to get channel message: ", err)
				}
				err = session.State.MessageAdd(loopTargetMsg)
				if err != nil {
					log.Panic("Failed to add message into state: ", err)
				}
			}
		}

		reverse(msgChain)
	}

	start := time.Now()

	response, err := chat.GptRequest(&msgChain, data, guild)
	if err != nil {
		if errors.As(err, &timeout) && timeout.Timeout() {
			ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.TimeOut)
			return
		} else {
			ErrorReply(session, orgMsg, err.Error())
			return
		}
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
