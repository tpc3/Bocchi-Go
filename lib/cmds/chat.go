package cmds

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	_ "golang.org/x/image/webp"

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
var detcost int

func ChatCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, msg *string, data *config.Tokens) {
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "🤔")

	if *msg == "" {
		ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.SubCmd)
		return
	}

	content, repnum, tmpnum, topnum, systemstr, model, cmodel, filter, imgurl, detail := splitMsg(msg, guild)

	if content == "" {
		ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.SubCmd)
		return
	}

	if strings.Contains(strings.ReplaceAll(*msg, content, ""), "-d") && !strings.Contains(strings.ReplaceAll(*msg, content, ""), "-i") {
		ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.NoImage)
		return
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

		start := time.Now()

		response, err := chat.GptRequest(&msgChain, data, guild, topnum, tmpnum, model, detcost)
		SendDiscord(session, orgMsg, guild, msg, data, response, err, start, filter, content, model)
		return
	}

	if imgurl != "" {

		re := regexp.MustCompile(`https?://[\w!?/+\-_~;.,*&@#$%()'[\]]+`)
		if !re.MatchString(imgurl) {
			ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.NoUrl)
			return
		}

		if detail == "miss" {
			ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.NoDetail)
			return
		}

		resp, err := http.Get(imgurl)
		if err != nil {
			if strings.Contains(err.Error(), "no such host") {
				ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.BrokenLink)
				return
			}
			log.Panic("Failed to get image: ", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.BrokenLink)
			return
		}
		imageConfig, imageType, err := image.DecodeConfig(resp.Body)
		if err != nil {
			if strings.Contains(err.Error(), "unknown format") {
				ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.NoSupportimage)
				return
			}
			log.Panic("Faild to decode: ", err)
		}

		if imageType != "png" && imageType != "jpeg" && imageType != "webp" && imageType != "gif" {
			ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.NoSupportimage)
			return
		}

		log.Print(imageType)

		model = "gpt-4-vision-preview"

		if filter {
			filter = false
		}

		img := []chat.Img{
			{
				Role: "user",
				Content: []chat.Content{
					chat.TextContent{
						Type: "text",
						Text: content,
					},
					chat.ImageContent{
						Type: "image_url",
						ImageURL: chat.ImageURL{
							Url:    imgurl,
							Detail: detail,
						},
					},
				},
			},
		}

		if detail == "low" {
			detcost = 85
		} else if detail == "high" {

			width := imageConfig.Width
			height := imageConfig.Height

			var newwidth, newheight int

			if width > 2048 || height > 2048 {
				if width > height {
					newwidth = 2048
					newheight = int(float64(height) * (2048.0 / float64(width)))
				} else {
					newheight = 2048
					newwidth = int(float64(width) * (2048.0 / float64(height)))
				}
			}
			if newwidth > 768 {
				newwidth = 768
				newheight = int(float64(height) * (768.0 / float64(newwidth)))
			} else if newheight > 768 {
				newheight = 768
				newwidth = int(float64(width) * (768.0 / float64(newheight)))
			} else {
				newwidth = width
				newheight = height
			}
			detcost = (((((newheight / 512) + 1) + ((newwidth / 512) + 1)) * 170) + 85)
		}

		start := time.Now()

		response, err := chat.GptRequestImg(&img, data, guild, topnum, tmpnum, model, detcost)

		SendDiscord(session, orgMsg, guild, msg, data, response, err, start, filter, content, model)
		return
	}

	detcost := 0

	var slog []string

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
			contentlog, _, _, _, systemstrlog, _, _, _, _, _ := splitMsg(&trimmed, guild)
			if systemstrlog != "" {
				slog = append(slog, systemstrlog)
			}
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
		}

		if systemstr == "" {
			if len(slog) != 0 {
				sysSlice := chat.Message{Role: "system", Content: slog[0]}
				msgChain = append([]chat.Message{sysSlice}, msgChain...)
			}
		} else {
			sysSlice := chat.Message{Role: "system", Content: systemstr}
			msgChain = append([]chat.Message{sysSlice}, msgChain...)
		}

		reverse(msgChain)
	} else {

		if systemstr != "" {
			sysSlice := chat.Message{Role: "system", Content: systemstr}
			msgChain = append([]chat.Message{sysSlice}, msgChain...)
		}
	}

	start := time.Now()

	response, err := chat.GptRequest(&msgChain, data, guild, topnum, tmpnum, model, detcost)

	SendDiscord(session, orgMsg, guild, msg, data, response, err, start, filter, content, model)

}

func SendDiscord(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, msg *string, data *config.Tokens, response string, err error, start time.Time, filter bool, content string, model string) {

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

func splitMsg(msg *string, guild *config.Guild) (string, int, float64, float64, string, string, bool, bool, string, string) {
	var content, systemstr, modelstr, imgurl, detail string
	var repnum int
	var tmpnum, topnum float64
	var prm, filter, cmodel bool

	modelstr, detail = guild.Model, "low"
	repnum, topnum, tmpnum = config.CurrentConfig.Guild.Reply, 1, 1
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
				i += 1
			} else if str[i] == "-l" {
				repnum, _ = strconv.Atoi(str[i+1])
				i += 1
			} else if str[i] == "-p" {
				topnum, _ = strconv.ParseFloat(str[i+1], 64)
				i += 1
			} else if str[i] == "-t" {
				tmpnum, _ = strconv.ParseFloat(str[i+1], 64)
				i += 1
			} else if str[i] == "-s" {
				systemstr = str[i+1]
				i += 1
			} else if str[i] == "-i" {
				imgurl = str[i+1]
				i += 1
			} else if str[i] == "-d" {
				if str[i+1] == "high" || str[i+1] == "low" {
					detail = str[i+1]
				} else {
					detail = "miss"
				}
				i += 1
			}
		} else {
			prm = false
			content += str[i] + " "
		}
	}
	return content, repnum, tmpnum, topnum, systemstr, modelstr, cmodel, filter, imgurl, detail
}
