package handler

import (
	"log"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tpc3/Bocchi-Go/lib/cmds"
	"github.com/tpc3/Bocchi-Go/lib/config"
	"github.com/tpc3/Bocchi-Go/lib/utils"
)

func MessageCreate(session *discordgo.Session, orgMsg *discordgo.MessageCreate) {
	defer func() {
		err := recover()
		if err != nil {
			log.Print("Oops, ", err)
			debug.PrintStack()
		}
	}()

	if config.CurrentConfig.Debug {
		start := time.Now()
		defer func() {
			log.Print("Message processed in ", time.Since(start).Milliseconds(), "ms.")
		}()
	}

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if orgMsg.Author.ID == session.State.User.ID || orgMsg.Content == "" {
		return
	}

	// Ignore all messages from blacklisted user
	for _, v := range config.CurrentConfig.UserBlacklist {
		if orgMsg.Author.ID == v {
			return
		}
	}

	// Ignore bot message
	if orgMsg.Author.Bot {
		return
	}
	prefix := config.CurrentConfig.Guild.Prefix
	guild := config.CurrentConfig.Guild
	date := config.CurrentData

	isCmd, trimmedMsg := utils.TrimPrefix(orgMsg.Content, prefix, session.State.User.Mention())

	if isCmd {
		if config.CurrentConfig.Debug {
			log.Print("Command processing")
		}
		cmds.HandleCmd(session, orgMsg, &guild, &date, &trimmedMsg)
		return
	}
}
