package cmds

import (
	"log"
	"runtime/debug"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/embed"
)

func ReplyEmbed(session *discordgo.Session, orgMsg *discordgo.MessageCreate, msgembed *discordgo.MessageEmbed) {
	reply := discordgo.MessageSend{}
	reply.Embed = msgembed
	if reply.Embed.Color == 0 {
		reply.Embed.Color = embed.ColorSystem
	}
	reply.Reference = orgMsg.Reference()
	reply.AllowedMentions = &discordgo.MessageAllowedMentions{
		RepliedUser: false,
	}
	_, err := session.ChannelMessageSendComplex(orgMsg.ChannelID, &reply)
	if err != nil {
		log.Print("Failed to send reply: ", err)
	}
}

func GPTReplyEmbed(session *discordgo.Session, orgMsg *discordgo.MessageCreate, msgembed *discordgo.MessageEmbed, model *string) {
	if strings.Contains(*model, "gpt-4") {
		msgembed.Color = embed.ColorGPT4
	} else {
		msgembed.Color = embed.ColorGPT3
	}
	ReplyEmbed(session, orgMsg, msgembed)
}

func UsageReply(session *discordgo.Session, orgMsg *discordgo.MessageCreate) {
	usage := embed.NewEmbed(session, orgMsg)
	usage.Fields = append(usage.Fields, &discordgo.MessageEmbedField{
		Name:  config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.ChatTitle,
		Value: config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.ChatUsage,
	})
	usage.Fields = append(usage.Fields, &discordgo.MessageEmbedField{
		Name:  config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.FilterTitle,
		Value: config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.FilterUsage,
	})
	usage.Fields = append(usage.Fields, &discordgo.MessageEmbedField{
		Name:  config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.PingTitle,
		Value: config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.PingUsage,
	})
	usage.Fields = append(usage.Fields, &discordgo.MessageEmbedField{
		Name:  config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.HelpTitle,
		Value: config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.HelpUsage,
	})
	usage.Fields = append(usage.Fields, &discordgo.MessageEmbedField{
		Name:  config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.ConfTitle,
		Value: config.Lang[config.CurrentConfig.Guild.Lang].Usage.Cmd.ConfUsage,
	})
	ReplyEmbed(session, orgMsg, usage)
}

func ErrorReply(session *discordgo.Session, orgMsg *discordgo.MessageCreate, description string) {
	msgEmbed := embed.NewEmbed(session, orgMsg)
	msgEmbed.Title = "Error"
	msgEmbed.Color = embed.ColorPink
	msgEmbed.Description = description
	ReplyEmbed(session, orgMsg, msgEmbed)
}

func UnknownError(session *discordgo.Session, orgMsg *discordgo.MessageCreate, lang *string, err error) {
	debug.PrintStack()
	msgEmbed := embed.NewEmbed(session, orgMsg)
	msgEmbed.Title = config.Lang[*lang].Error.UnknownTitle
	msgEmbed.Description = config.Lang[*lang].Error.UnknownDesc + "\n`" + err.Error() + "`"
	msgEmbed.Color = embed.ColorPink
	ReplyEmbed(session, orgMsg, msgEmbed)
}

func HandleCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, data *config.Tokens, message *string) {
	splitMsg := strings.SplitN(*message, " ", 2)
	var param string
	if len(splitMsg) == 2 {
		param = splitMsg[1]
	} else {
		param = ""
	}
	switch splitMsg[0] {
	case Ping:
		PingCmd(session, orgMsg)
	case Help:
		HelpCmd(session, orgMsg)
	case Config:
		ConfigCmd(session, orgMsg, *guild, &param)
	case Cost:
		CostCmd(session, orgMsg, guild)
	case Chat:
		go ChatCmd(session, orgMsg, guild, &param, data)
	default:
		ErrorReply(session, orgMsg, config.Lang[guild.Lang].Error.NoCmd)
	}
}
