package cmds

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/embed"
)

const Help = "help"

func HelpCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, message *string) {
	log.Print(guild)
	msg := embed.NewEmbed(session, orgMsg)
	msg.Title = "Help"
	msg.Description = config.Lang[guild.Lang].Help + "\n" + config.CurrentConfig.Help
	ReplyEmbed(session, orgMsg, msg)
}
