package discord

import "github.com/bwmarrin/discordgo"

func (r *Router) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if r == nil || r.anonymous == nil {
		return
	}
	r.anonymous.HandleMessageCreate(s, m)
}
