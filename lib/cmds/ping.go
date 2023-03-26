package cmds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/embed"
)

const Ping = "ping"

func PingCmd(session *discordgo.Session, orgMsg *discordgo.MessageCreate, guild *config.Guild, message *string) {
	embedMsg := embed.NewEmbed(session, orgMsg)
	embedMsg.Title = "Pong!"
	ReplyEmbed(session, orgMsg, embedMsg)
	session.MessageReactionAdd(orgMsg.ChannelID, orgMsg.ID, "🏓")
}
