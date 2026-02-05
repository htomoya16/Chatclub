-- Create "sf6_accounts" table
CREATE TABLE "public"."sf6_accounts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "guild_id" text NOT NULL,
  "user_id" text NOT NULL,
  "fighter_id" text NOT NULL,
  "display_name" text NULL,
  "status" text NOT NULL DEFAULT 'active',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sf6_accounts_guild_fighter_unique" UNIQUE ("guild_id", "fighter_id"),
  CONSTRAINT "sf6_accounts_guild_user_unique" UNIQUE ("guild_id", "user_id"),
  CONSTRAINT "sf6_accounts_guild_id_fkey" FOREIGN KEY ("guild_id") REFERENCES "public"."guilds" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sf6_accounts_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sf6_accounts_status_check" CHECK (status = ANY (ARRAY['active'::text, 'inactive'::text]))
);
-- Create "sf6_sessions" table
CREATE TABLE "public"."sf6_sessions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "guild_id" text NOT NULL,
  "user_id" text NOT NULL,
  "opponent_fighter_id" text NOT NULL,
  "status" text NOT NULL DEFAULT 'active',
  "started_at" timestamptz NOT NULL,
  "ended_at" timestamptz NULL,
  "last_polled_at" timestamptz NOT NULL,
  "last_seen_battle_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sf6_sessions_guild_id_fkey" FOREIGN KEY ("guild_id") REFERENCES "public"."guilds" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sf6_sessions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sf6_sessions_status_check" CHECK (status = ANY (ARRAY['active'::text, 'ended'::text]))
);
-- Create index "sf6_sessions_guild_user_opponent_idx" to table: "sf6_sessions"
CREATE INDEX "sf6_sessions_guild_user_opponent_idx" ON "public"."sf6_sessions" ("guild_id", "user_id", "opponent_fighter_id");
-- Create index "sf6_sessions_guild_user_status_idx" to table: "sf6_sessions"
CREATE INDEX "sf6_sessions_guild_user_status_idx" ON "public"."sf6_sessions" ("guild_id", "user_id", "status");
-- Create "sf6_battles" table
CREATE TABLE "public"."sf6_battles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "guild_id" text NOT NULL,
  "user_id" text NOT NULL,
  "opponent_fighter_id" text NOT NULL,
  "battle_at" timestamptz NOT NULL,
  "result" text NOT NULL,
  "self_character" text NOT NULL,
  "opponent_character" text NOT NULL,
  "round_wins" integer NOT NULL,
  "round_losses" integer NOT NULL,
  "source_battle_id" text NULL,
  "source_key" text NOT NULL,
  "session_id" uuid NULL,
  "raw_payload" jsonb NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sf6_battles_unique" UNIQUE ("guild_id", "user_id", "source_key"),
  CONSTRAINT "sf6_battles_guild_id_fkey" FOREIGN KEY ("guild_id") REFERENCES "public"."guilds" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sf6_battles_session_id_fkey" FOREIGN KEY ("session_id") REFERENCES "public"."sf6_sessions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sf6_battles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sf6_battles_result_check" CHECK (result = ANY (ARRAY['win'::text, 'loss'::text, 'draw'::text]))
);
-- Create index "sf6_battles_guild_user_battle_at_idx" to table: "sf6_battles"
CREATE INDEX "sf6_battles_guild_user_battle_at_idx" ON "public"."sf6_battles" ("guild_id", "user_id", "battle_at");
-- Create index "sf6_battles_guild_user_opponent_battle_at_idx" to table: "sf6_battles"
CREATE INDEX "sf6_battles_guild_user_opponent_battle_at_idx" ON "public"."sf6_battles" ("guild_id", "user_id", "opponent_fighter_id", "battle_at");
-- Create "sf6_friends" table
CREATE TABLE "public"."sf6_friends" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "guild_id" text NOT NULL,
  "user_id" text NOT NULL,
  "fighter_id" text NOT NULL,
  "display_name" text NULL,
  "alias" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sf6_friends_unique" UNIQUE ("guild_id", "user_id", "fighter_id"),
  CONSTRAINT "sf6_friends_guild_id_fkey" FOREIGN KEY ("guild_id") REFERENCES "public"."guilds" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sf6_friends_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
