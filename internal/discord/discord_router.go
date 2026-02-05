package discord

import (
	"backend/internal/service"

	"github.com/bwmarrin/discordgo"
)

// Router は Discord の Interaction を各ハンドラに振り分ける役割。
type Router struct {
	AnonymousChannelService service.AnonymousChannelService
	SF6AccountService       service.SF6AccountService
	SF6FriendService        service.SF6FriendService
	SF6Service              service.SF6Service
	SF6SessionService       service.SF6SessionService
	// TournamentService service.TournamentService
	// CypherService     service.CypherService
	// BeatService       service.BeatService
}

// NewRouter で必要な service を全部 DI しておく。
func NewRouter(
	anonymousChannelService service.AnonymousChannelService,
	sf6AccountService service.SF6AccountService,
	sf6FriendService service.SF6FriendService,
	sf6Service service.SF6Service,
	sf6SessionService service.SF6SessionService,
	// tournamentService service.TournamentService,
	// cypherService service.CypherService,
	// beatService service.BeatService,
) *Router {
	return &Router{
		AnonymousChannelService: anonymousChannelService,
		SF6AccountService:       sf6AccountService,
		SF6FriendService:        sf6FriendService,
		SF6Service:              sf6Service,
		SF6SessionService:       sf6SessionService,
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
		case "sf6_stats":
			r.handleSF6Stats(s, i)
		case "sf6_history":
			r.handleSF6History(s, i)
		case "sf6_session":
			r.handleSF6Session(s, i)
		case "sf6_friend":
			r.handleSF6Friend(s, i)

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
