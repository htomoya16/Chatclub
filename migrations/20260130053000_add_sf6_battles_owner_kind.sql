-- Add owner_kind to sf6_battles
ALTER TABLE "public"."sf6_battles"
  ADD COLUMN "owner_kind" text NOT NULL DEFAULT 'account',
  ADD CONSTRAINT "sf6_battles_owner_kind_check" CHECK ("owner_kind" IN ('account','friend'));
