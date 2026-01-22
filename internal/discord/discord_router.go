package discord

import (
	"backend/internal/service"

	"github.com/bwmarrin/discordgo"
)

// Router は Discord の Interaction を各ハンドラに振り分ける役割。
type Router struct {
	WhitelistService service.WhitelistService
	// TournamentService service.TournamentService
	// CypherService     service.CypherService
	// BeatService       service.BeatService
}

// NewRouter で必要な service を全部 DI しておく。
// まだトーナメント等が未実装なら WhitelistService だけでもOK。
func NewRouter(
	whitelistService service.WhitelistService,
	// tournamentService service.TournamentService,
	// cypherService service.CypherService,
	// beatService service.BeatService,
) *Router {
	return &Router{
		WhitelistService: whitelistService,
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
