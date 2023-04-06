package cmds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/embed"
)

const Help = "help"

func HelpCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate) {
	msg := embed.NewEmbed(session, orgMsg)
	msg.Title = "Help"
	UsageReply(session, orgMsg)
}
