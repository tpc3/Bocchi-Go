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

const Chat = "chat"

func ChatCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, param *string) {
	msg := *param
	if msg == "" {
		ErrorReply(session, orgMsg, config.Lang[config.CurrentConfig.Guild.Lang].Error.SubCmd)
		return
	}
	start := time.Now()
	go func() {
		session.ChannelTyping(orgMsg.ChannelID)
	}()
	response, coststr := chat.GptRequest(guild, &msg)
	if utf8.RuneCountInString(response) > 4096 {
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
	GPTReplyEmbed(session, orgMsg, embedMsg)
}
