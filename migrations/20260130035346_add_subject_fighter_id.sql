-- Modify "sf6_battles" table
ALTER TABLE "public"."sf6_battles" DROP CONSTRAINT "sf6_battles_unique", ADD COLUMN "subject_fighter_id" text NOT NULL, ADD CONSTRAINT "sf6_battles_unique" UNIQUE ("guild_id", "subject_fighter_id", "source_key");
