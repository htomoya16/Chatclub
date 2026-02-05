-- Extend owner_kind allowed values to include 'unlinked'
ALTER TABLE "public"."sf6_battles"
  DROP CONSTRAINT "sf6_battles_owner_kind_check",
  ADD CONSTRAINT "sf6_battles_owner_kind_check" CHECK ("owner_kind" IN ('account','friend','unlinked'));
