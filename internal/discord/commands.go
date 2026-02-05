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
			Description: "Fetch and store SF6 battle log (admin only).",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
		},
		{
			Name:        "sf6_stats",
			Description: "Show SF6 stats (range/count).",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "range",
					Description: "Stats by date range (JST).",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "opponent_code",
							Description: "Opponent SF6 user code (sid)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "from",
							Description: "Start date (YYYY-MM-DD, JST)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "to",
							Description: "End date (YYYY-MM-DD, JST)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subject_code",
							Description: "Subject SF6 user code (sid)",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "count",
					Description: "Stats by recent N matches.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "opponent_code",
							Description: "Opponent SF6 user code (sid)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "count",
							Description: "Recent match count",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subject_code",
							Description: "Subject SF6 user code (sid)",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "Stats grouped by <=30min intervals.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "opponent_code",
							Description: "Opponent SF6 user code (sid)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subject_code",
							Description: "Subject SF6 user code (sid)",
							Required:    false,
						},
					},
				},
			},
		},
		{
			Name:        "sf6_history",
			Description: "Show SF6 battle history.",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "opponent_code",
					Description: "Opponent SF6 user code (sid)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "subject_code",
					Description: "Subject SF6 user code (sid)",
					Required:    false,
				},
			},
		},
		{
			Name:        "sf6_session",
			Description: "Start/end a session and show stats.",
			DMPermission: func() *bool {
				v := false
				return &v
			}(),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "start",
					Description: "Start a session.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "opponent_code",
							Description: "Opponent SF6 user code (sid)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subject_code",
							Description: "Subject SF6 user code (sid)",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "end",
					Description: "End a session and show stats.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "opponent_code",
							Description: "Opponent SF6 user code (sid)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subject_code",
							Description: "Subject SF6 user code (sid)",
							Required:    false,
						},
					},
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
			Name:        "sf6_friend",
			Description: "Show SF6 friend list and controls.",
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
