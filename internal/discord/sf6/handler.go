package sf6

import (
	"backend/internal/service"

	"github.com/bwmarrin/discordgo"
)

type Handler struct {
	SF6AccountService service.SF6AccountService
	SF6FriendService  service.SF6FriendService
	SF6Service        service.SF6Service
	SF6SessionService service.SF6SessionService
}

func NewHandler(
	sf6AccountService service.SF6AccountService,
	sf6FriendService service.SF6FriendService,
	sf6Service service.SF6Service,
	sf6SessionService service.SF6SessionService,
) *Handler {
	return &Handler{
		SF6AccountService: sf6AccountService,
		SF6FriendService:  sf6FriendService,
		SF6Service:        sf6Service,
		SF6SessionService: sf6SessionService,
	}
}

func (h *Handler) HandleAccount(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6Link(s, i)
}

func (h *Handler) HandleUnlink(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6Unlink(s, i)
}

func (h *Handler) HandleFetch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6Fetch(s, i)
}

func (h *Handler) HandleStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6Stats(s, i)
}

func (h *Handler) HandleHistory(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6History(s, i)
}

func (h *Handler) HandleSession(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6Session(s, i)
}

func (h *Handler) HandleFriend(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6Friend(s, i)
}

func (h *Handler) HandleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6Component(s, i)
}

func (h *Handler) HandleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.handleSF6ModalSubmit(s, i)
}
