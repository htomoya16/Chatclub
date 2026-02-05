package discord

import (
	"backend/internal/discord/anonymous"
	"backend/internal/discord/sf6"
	"backend/internal/service"

	"github.com/bwmarrin/discordgo"
)

// Router は Discord の Interaction を各ハンドラに振り分ける役割。
type Router struct {
	anonymous *anonymous.Handler
	sf6       *sf6.Handler
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
		anonymous: anonymous.NewHandler(anonymousChannelService),
		sf6:       sf6.NewHandler(sf6AccountService, sf6FriendService, sf6Service, sf6SessionService),
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
			r.anonymous.HandleAnon(s, i)
		case "anon-channel":
			r.anonymous.HandleAnonChannel(s, i)
		case "sf6_account":
			r.sf6.HandleAccount(s, i)
		case "sf6_unlink":
			r.sf6.HandleUnlink(s, i)
		case "sf6_fetch":
			r.sf6.HandleFetch(s, i)
		case "sf6_stats":
			r.sf6.HandleStats(s, i)
		case "sf6_history":
			r.sf6.HandleHistory(s, i)
		case "sf6_session":
			r.sf6.HandleSession(s, i)
		case "sf6_friend":
			r.sf6.HandleFriend(s, i)

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
		r.sf6.HandleComponent(s, i)
	case discordgo.InteractionModalSubmit:
		r.sf6.HandleModalSubmit(s, i)
	default:
		return
	}
}
