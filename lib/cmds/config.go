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
	ReplyEmbed(session, orgMsg, msg)
}

func ConfigCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild config.Guild, message *string) {
	maxtoken := strconv.Itoa(guild.MaxToken)
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
			Name:  "maxtoken",
			Value: maxtoken,
		})
		ReplyEmbed(session, orgMsg, msg)
		return
	}
	if len(split) != 2 {
		ConfigUsage(session, orgMsg, &guild, errors.New("not enough arguments"))
		return
	}
	ok := false
	switch split[0] {
	case "prefix":
		guild.Prefix = split[1]
	case "lang":
		_, ok = config.Lang[split[1]]
		if ok {
			guild.Lang = split[1]
		} else {
			ErrorReply(session, orgMsg, "unsupported language")
			return
		}
	case "maxtoken":
		maxtoken = split[1]
		guild.MaxToken, _ = strconv.Atoi(maxtoken)
		if guild.MaxToken < 1 || guild.MaxToken > 4095 {
			ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.MustValue)
			return
		}
	default:
		ConfigUsage(session, orgMsg, &guild, errors.New("item not found"))
		return
	}
	err := config.VerifyGuild(&guild)
	if err != nil {
		ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.SubCmd)
		return
	}
	err = config.SaveGuild(&orgMsg.GuildID, &guild)
	if err != nil {
		UnknownError(session, orgMsg, &guild.Lang, err)
		return
	}
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "👍")
}
