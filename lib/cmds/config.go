package cmds

import (
	"errors"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/embed"
)

const (
	Config     = "config"
	configFile = "./config.yml"
)

func ConfigUsage(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, err error) {
	msg := embed.NewEmbed(session, orgMsg)
	if err != nil {
		msg.Title = config.Lang[guild.Lang].Error.Syntax
		msg.Description = err.Error() + "\n"
		msg.Color = embed.ColorPink
	}
	msg.Description += "`" + guild.Prefix + Config + " [<item> <value>]`\n" + config.Lang[guild.Lang].Usage.Config.Desc
	msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
		Name:  "prefix <prefix>",
		Value: config.Lang[guild.Lang].Usage.Config.Prefix,
	})
	msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
		Name:  "lang <language>",
		Value: config.Lang[guild.Lang].Usage.Config.Lang,
	})
	msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
		Name:  "model <model name>",
		Value: config.Lang[guild.Lang].Usage.Config.Model,
	})
	msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
		Name:  "timeout <int>",
		Value: config.Lang[guild.Lang].Usage.Config.TimeOut,
	})
	msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
		Name:  "reply <int>",
		Value: config.Lang[guild.Lang].Usage.Config.Reply,
	})
	ReplyEmbed(session, orgMsg, msg)
}

func ConfigCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild config.Guild, message *string) {
	split := strings.SplitN(*message, " ", 2)
	if *message == "" {
		msg := embed.NewEmbed(session, orgMsg)
		msg.Title = config.Lang[guild.Lang].CurrConf
		msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
			Name:  "prefix",
			Value: guild.Prefix,
		})
		msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
			Name:  "lang",
			Value: guild.Lang,
		})
		msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
			Name:  "model",
			Value: guild.Model,
		})
		msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
			Name:  "timeout",
			Value: strconv.Itoa(guild.Timeout),
		})
		msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
			Name:  "reply",
			Value: strconv.Itoa(guild.Reply),
		})
		ReplyEmbed(session, orgMsg, msg)
		return
	}
	if len(split) != 2 {
		ConfigUsage(session, orgMsg, &guild, errors.New("not enough arguments"))
		return
	}
	ok := false
	var key, item string
	switch split[0] {
	case "prefix":
		guild.Prefix = split[1]
		key = config.Lang[guild.Lang].Config.Item.Prefix
		item = guild.Prefix
	case "lang":
		_, ok = config.Lang[split[1]]
		if ok {
			guild.Lang = split[1]
			key = config.Lang[guild.Lang].Config.Item.Lang
			item = guild.Lang
		} else {
			ErrorReply(session, orgMsg, "unsupported language")
			return
		}
	case "model":
		if split[1] != "gpt-4" && split[1] != "gpt-4-32k" && split[1] != "gpt-3.5-turbo" {
			ErrorReply(session, orgMsg, "unsupported models")
			return
		} else {
			guild.Model = split[1]
			key = config.Lang[guild.Lang].Config.Item.Model
			item = guild.Model
		}
	case "timeout":
		timeout := split[1]
		guild.Timeout, _ = strconv.Atoi(timeout)
		if guild.Timeout < 1 {
			ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.MustTimeoutDuration)
			return
		}
		key = config.Lang[guild.Lang].Config.Item.Timeout
		item = timeout
	case "reply":
		reply := split[1]
		guild.Reply, _ = strconv.Atoi(reply)
		if guild.Reply < 1 {
			ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.MustTimeoutDuration)
			return
		}
		guild.Reply, _ = strconv.Atoi(reply)
		key = config.Lang[guild.Lang].Config.Item.Reply
		item = reply
	default:
		ConfigUsage(session, orgMsg, &guild, errors.New("item not found"))
		return
	}
	err := config.VerifyGuild(&guild)
	if err != nil {
		ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.SubCmd)
		return
	}
	err = config.SaveGuild(&guild)
	if err != nil {
		UnknownError(session, orgMsg, &guild.Lang, err)
		return
	}
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "👍")
	msg := embed.NewEmbed(session, orgMsg)
	msg.Title = config.Lang[guild.Lang].Config.Title
	msg.Color = embed.ColorSystem
	msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
		Value: key + item + config.Lang[guild.Lang].Config.Announce,
	})
	ReplyEmbed(session, orgMsg, msg)
}
