package discord

import "github.com/bwmarrin/discordgo"

// Commands はこのBotで使う全てのスラッシュコマンド定義を返す。
func Commands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Check if the bot is alive.",
		},
		// ここに今後 /tournament /beat /cypher を足していく:
		// {
		// 	Name:        "tournament",
		// 	Description: "Tournament operations",
		// 	Options: []*discordgo.ApplicationCommandOption{
		// 		{
		// 			Type:        discordgo.ApplicationCommandOptionSubCommand,
		// 			Name:        "create",
		// 			Description: "Create a new tournament",
		// 		},
		// 		// ...
		// 	},
		// },
	}
}
