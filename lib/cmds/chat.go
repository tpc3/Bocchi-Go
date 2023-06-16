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
	Chat = "chat"
)

var timeout *url.Error

func ChatCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, msg *string, data *config.Tokens) {
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "🤔")
	if *msg == "" {
		ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.SubCmd)
		return
	}

	content, repnum, tmpnum, topnum, systemstr, model, filter := splitMsg(msg, guild)

	msgChain := []chat.Message{{Role: "user", Content: content}}

	if filter {
		if orgMsg.ReferencedMessage != nil {
			filterMsg, err := session.State.Message(orgMsg.ChannelID, orgMsg.ReferencedMessage.ID)
			if err != nil {
				filterMsg, err = session.ChannelMessage(orgMsg.ChannelID, orgMsg.ReferencedMessage.ID)
				if err != nil {
					log.Panic("Failed to get channel message: ", err)
				}
				err = session.State.MessageAdd(filterMsg)
				if err != nil {
					log.Panic("Failed to add message into state: ", err)
				}
			}

			if !filterMsg.Author.Bot {
				msgChain = []chat.Message{{Role: "user", Content: filterMsg.Content + "\n\n以上の文章をポジティブな言葉で言い換えてください。"}}
				topnum, tmpnum, model = 1, 1, "gpt-3.5-turbo"
			}
		} else {
			msgChain = []chat.Message{{Role: "user", Content: content + "\n\n以上の文章をポジティブな言葉で言い換えてください。"}}
			topnum, tmpnum, model = 1, 1, "gpt-3.5-turbo"
		}
	} else {

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
			for i := 0; i < repnum; i++ {
				if loopTargetMsg.Author.ID != session.State.User.ID {
					break
				} else if loopTargetMsg.Embeds[0].Color != embed.ColorGPT3 && loopTargetMsg.Embeds[0].Color != embed.ColorGPT4 { //ToDo: Better handling
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

				if loopTargetMsg.Content == "" {
					break
				}
				_, trimmed := utils.TrimPrefix(loopTargetMsg.Content, config.CurrentConfig.Guild.Prefix+Chat+" ", session.State.User.Mention())
				content, _, tmpnum, topnum, systemstr, model, filter = splitMsg(&trimmed, guild)
				msgChain = append(msgChain, chat.Message{Role: "user", Content: content})
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

		if systemstr != "" {
			sysSlice := chat.Message{Role: "system", Content: systemstr}
			msgChain = append([]chat.Message{sysSlice}, msgChain...)
		}
	}

	start := time.Now()

	response, err := chat.GptRequest(&msgChain, data, guild, topnum, tmpnum, model)
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
	if filter {
		embedMsg.Title = "Social Filter"
	} else if utf8.RuneCountInString(content) > 50 {
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
	GPTReplyEmbed(session, orgMsg, embedMsg, &model)
}

// https://stackoverflow.com/questions/28058278/how-do-i-reverse-a-slice-in-go
func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func splitMsg(msg *string, guild *config.Guild) (string, int, float64, float64, string, string, bool) {
	var content, systemstr, modelstr string
	var repnum int
	var tmpnum, topnum float64
	var filter bool

	modelstr = guild.Model
	repnum, topnum, tmpnum = 1, 1, 1
	filter = false

	str := strings.Split(*msg, " ")
	for i, word := range str {
		if word == "-l" && i+1 < len(str) {
			value, err := strconv.Atoi(str[i+1])
			if err == nil && value > 1 {
				repnum = value
			}
		} else if word == "-p" && i+1 < len(str) {
			value, err := strconv.ParseFloat(str[i+1], 64)
			if err == nil && 0 <= value && value <= 1 {
				topnum = value
			}
		} else if word == "-t" && i+1 < len(str) {
			value, err := strconv.ParseFloat(str[i+1], 64)
			if err == nil && 0 <= value && value <= 2 {
				tmpnum = value
			}
		} else if word == "-s" && i+1 < len(str) {
			systemstr = str[i+1]
		} else if word == "-m" && i+1 < len(str) {
			modelstr = str[i+1]
		} else if word == "-f" {
			filter = true
		} else if !strings.HasPrefix(word, "-") {
			if filter {
				content += word
			} else if i == 0 {
				content += word + " "
			} else if !strings.HasPrefix(str[i-1], "-") {
				content += word + " "
			}
		}
	}
	return content, repnum, tmpnum, topnum, systemstr, modelstr, filter
}
