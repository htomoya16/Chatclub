-- Create "guilds" table
CREATE TABLE "public"."guilds" (
  "id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create "anonymous_channels" table
CREATE TABLE "public"."anonymous_channels" (
  "guild_id" text NOT NULL,
  "channel_id" text NOT NULL,
  "webhook_id" text NOT NULL,
  "webhook_token" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("guild_id", "channel_id"),
  CONSTRAINT "anonymous_channels_guild_id_fkey" FOREIGN KEY ("guild_id") REFERENCES "public"."guilds" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" text NOT NULL,
  "is_bot" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create "guild_members" table
CREATE TABLE "public"."guild_members" (
  "guild_id" text NOT NULL,
  "user_id" text NOT NULL,
  "joined_at" timestamptz NULL,
  "left_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("guild_id", "user_id"),
  CONSTRAINT "guild_members_guild_id_fkey" FOREIGN KEY ("guild_id") REFERENCES "public"."guilds" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "guild_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
