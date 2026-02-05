-- Core tables: guilds/users/members
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS guilds (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    is_bot BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS guild_members (
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    joined_at TIMESTAMPTZ,
    left_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (guild_id, user_id),
    FOREIGN KEY (guild_id) REFERENCES guilds (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- Anonymous chat: channel settings
CREATE TABLE IF NOT EXISTS anonymous_channels (
    guild_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    webhook_id TEXT NOT NULL,
    webhook_token TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (guild_id, channel_id),
    FOREIGN KEY (guild_id) REFERENCES guilds (id) ON DELETE CASCADE
);

-- SF6 Buckler: accounts
CREATE TABLE IF NOT EXISTS sf6_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    fighter_id TEXT NOT NULL,
    display_name TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sf6_accounts_status_check CHECK (status IN ('active','inactive')),
    CONSTRAINT sf6_accounts_guild_user_unique UNIQUE (guild_id, user_id),
    CONSTRAINT sf6_accounts_guild_fighter_unique UNIQUE (guild_id, fighter_id),
    FOREIGN KEY (guild_id) REFERENCES guilds (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- SF6 Buckler: friends
CREATE TABLE IF NOT EXISTS sf6_friends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    fighter_id TEXT NOT NULL,
    display_name TEXT,
    alias TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sf6_friends_unique UNIQUE (guild_id, user_id, fighter_id),
    FOREIGN KEY (guild_id) REFERENCES guilds (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- SF6 Buckler: sessions
CREATE TABLE IF NOT EXISTS sf6_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    opponent_fighter_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    last_polled_at TIMESTAMPTZ NOT NULL,
    last_seen_battle_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sf6_sessions_status_check CHECK (status IN ('active','ended')),
    FOREIGN KEY (guild_id) REFERENCES guilds (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS sf6_sessions_guild_user_status_idx
    ON sf6_sessions (guild_id, user_id, status);
CREATE INDEX IF NOT EXISTS sf6_sessions_guild_user_opponent_idx
    ON sf6_sessions (guild_id, user_id, opponent_fighter_id);

-- SF6 Buckler: battles
CREATE TABLE IF NOT EXISTS sf6_battles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    owner_kind TEXT NOT NULL DEFAULT 'account',
    subject_fighter_id TEXT NOT NULL,
    opponent_fighter_id TEXT NOT NULL,
    battle_at TIMESTAMPTZ NOT NULL,
    result TEXT NOT NULL,
    self_character TEXT NOT NULL,
    opponent_character TEXT NOT NULL,
    round_wins INT NOT NULL,
    round_losses INT NOT NULL,
    source_key TEXT NOT NULL,
    session_id UUID,
    raw_payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sf6_battles_result_check CHECK (result IN ('win','loss','draw')),
    CONSTRAINT sf6_battles_owner_kind_check CHECK (owner_kind IN ('account','friend','unlinked')),
    CONSTRAINT sf6_battles_unique UNIQUE (guild_id, subject_fighter_id, source_key),
    FOREIGN KEY (session_id) REFERENCES sf6_sessions (id) ON DELETE SET NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS sf6_battles_guild_user_opponent_battle_at_idx
    ON sf6_battles (guild_id, user_id, opponent_fighter_id, battle_at);
CREATE INDEX IF NOT EXISTS sf6_battles_guild_user_battle_at_idx
    ON sf6_battles (guild_id, user_id, battle_at);
