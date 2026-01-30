package discord

import (
	"backend/internal/service"

	"github.com/bwmarrin/discordgo"
)

// Router は Discord の Interaction を各ハンドラに振り分ける役割。
type Router struct {
	AnonymousChannelService service.AnonymousChannelService
	SF6AccountService       service.SF6AccountService
	SF6Service              service.SF6Service
	// TournamentService service.TournamentService
	// CypherService     service.CypherService
	// BeatService       service.BeatService
}

// NewRouter で必要な service を全部 DI しておく。
func NewRouter(
	anonymousChannelService service.AnonymousChannelService,
	sf6AccountService service.SF6AccountService,
	sf6Service service.SF6Service,
	// tournamentService service.TournamentService,
	// cypherService service.CypherService,
	// beatService service.BeatService,
) *Router {
	return &Router{
		AnonymousChannelService: anonymousChannelService,
		SF6AccountService:       sf6AccountService,
		SF6Service:              sf6Service,
		// TournamentService: tournamentService,
		// CypherService:     cypherService,
		// BeatService:       beatService,
	}
}

// HandleInteraction は discordgo のイベントハンドラとして登録される入口。
// main.go 側で session.AddHandler(router.HandleInteraction) する想定。
func (r *Router) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		switch data.Name {
		case "ping":
			// /ping
			r.handlePing(s, i)
		case "anon":
			r.handleAnon(s, i)
		case "anon-channel":
			r.handleAnonChannel(s, i)
		case "sf6_account":
			r.handleSF6Link(s, i)
		case "sf6_unlink":
			r.handleSF6Unlink(s, i)
		case "sf6_fetch":
			r.handleSF6Fetch(s, i)

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
	case discordgo.InteractionMessageComponent:
		r.handleSF6Component(s, i)
	case discordgo.InteractionModalSubmit:
		r.handleSF6ModalSubmit(s, i)
	default:
		return
	}
}
