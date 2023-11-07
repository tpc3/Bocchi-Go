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

	content, repnum, tmpnum, topnum, systemstr, model, cmodel, filter := splitMsg(msg, guild)

	if content == "" {
		ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.SubCmd)
	}

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

				if i == 0 && !cmodel {
					if loopTargetMsg.Embeds[0].Color == embed.ColorGPT3 {
						model = "gpt-3.5-turbo"
					} else if loopTargetMsg.Embeds[0].Color == embed.ColorGPT4 {
						model = "gpt-4"
					} else {
						ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.CantReply)
					}
				}

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
				contentlog, _, _, _, systemstrlog, _, _, _ := splitMsg(&trimmed, guild)
				msgChain = append(msgChain, chat.Message{Role: "user", Content: contentlog})
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

				if systemstr != "" || systemstrlog != "" {
					if systemstrlog != "" {
						sysSlice := chat.Message{Role: "system", Content: systemstrlog}
						msgChain = append([]chat.Message{sysSlice}, msgChain...)
					} else {
						sysSlice := chat.Message{Role: "system", Content: systemstr}
						msgChain = append([]chat.Message{sysSlice}, msgChain...)
					}
				}
			}

			reverse(msgChain)
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

func splitMsg(msg *string, guild *config.Guild) (string, int, float64, float64, string, string, bool, bool) {
	var content, systemstr, modelstr string
	var repnum int
	var tmpnum, topnum float64
	var prm, filter, cmodel bool

	modelstr = guild.Model
	repnum, topnum, tmpnum = 1, 1, 1
	prm, filter, cmodel = true, false, false

	str := strings.Split(*msg, " ")
	leng := len(str)

	for i := 0; i < leng; i++ {
		if strings.HasPrefix(str[i], "-") && prm && !filter {
			if str[i] == "-f" {
				filter = true
			} else if str[i] == "-m" {
				modelstr = str[i+1]
				cmodel = true
			} else if str[i] == "-l" {
				repnum, _ = strconv.Atoi(str[i+1])
			} else if str[i] == "-p" {
				topnum, _ = strconv.ParseFloat(str[i+1], 64)
			} else if str[i] == "-t" {
				tmpnum, _ = strconv.ParseFloat(str[i+1], 64)
			} else if str[i] == "-s" {
				systemstr = str[i+1]
			}
			i += 1
		} else {
			prm = false
			content += str[i] + " "
		}
	}
	return content, repnum, tmpnum, topnum, systemstr, modelstr, cmodel, filter
}
