package discord

import (
	"backend/internal/service"

	"github.com/bwmarrin/discordgo"
)

// Router は Discord の Interaction を各ハンドラに振り分ける役割。
type Router struct {
	AnonymousChannelService service.AnonymousChannelService
	// TournamentService service.TournamentService
	// CypherService     service.CypherService
	// BeatService       service.BeatService
}

// NewRouter で必要な service を全部 DI しておく。
func NewRouter(
	anonymousChannelService service.AnonymousChannelService,
	// tournamentService service.TournamentService,
	// cypherService service.CypherService,
	// beatService service.BeatService,
) *Router {
	return &Router{
		AnonymousChannelService: anonymousChannelService,
		// TournamentService: tournamentService,
		// CypherService:     cypherService,
		// BeatService:       beatService,
	}
}

// HandleInteraction は discordgo のイベントハンドラとして登録される入口。
// main.go 側で session.AddHandler(router.HandleInteraction) する想定。
func (r *Router) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Slash Command 以外は今は無視
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	switch data.Name {
	case "ping":
		// /ping
		r.handlePing(s, i)
	case "anon":
		r.handleAnon(s, i)
	case "anon-channel":
		r.handleAnonChannel(s, i)

	// 将来的な拡張 (コメントアウトしておいてOK)
	// case "tournament":
	// 	r.handleTournament(s, i)
	// case "cypher":
	// 	r.handleCypher(s, i)
	// case "beat":
	// 	r.handleBeat(s, i)
	default:
		// 未対応コマンドはとりあえず無視 or ログに出すくらいでOK
		return
	}
}
