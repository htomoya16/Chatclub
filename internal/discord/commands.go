package discord

import "github.com/bwmarrin/discordgo"

// Commands はこのBotで使う全てのスラッシュコマンド定義を返す。
func Commands() []*discordgo.ApplicationCommand {
	manageChannelsPerm := int64(discordgo.PermissionManageChannels)
	return []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Check if the bot is alive.",
		},
		{
			Name:        "sf6_account",
			Description: "Show SF6 account status and controls.",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
		},
		{
			Name:        "sf6_fetch",
			Description: "Fetch and store SF6 battle log.",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user_code",
					Description: "SF6 user code (sid)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "Page number (optional)",
					Required:    false,
				},
			},
		},
		{
			Name:        "sf6_unlink",
			Description: "Unlink your SF6 account.",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
		},
		{
			Name:        "anon",
			Description: "Post anonymously in this channel.",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Message content",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "file1",
					Description: "Attachment 1",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "file2",
					Description: "Attachment 2",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "file3",
					Description: "Attachment 3",
					Required:    false,
				},
			},
		},
		{
			Name:        "anon-channel",
			Description: "Manage anonymous channels",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
			DefaultMemberPermissions: &manageChannelsPerm,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Add a channel to anonymous list",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "channel",
							Description: "Target text channel",
							Required:    true,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildText,
								discordgo.ChannelTypeGuildNews,
							},
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "Remove a channel from anonymous list",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "channel",
							Description: "Target text channel",
							Required:    true,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildText,
								discordgo.ChannelTypeGuildNews,
							},
						},
					},
				},
			},
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
