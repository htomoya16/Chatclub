-- Create "servers" table
CREATE TABLE `servers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `discord_guild_id` varchar(32) NOT NULL,
  `name` varchar(191) NULL,
  `icon_url` varchar(255) NULL,
  `owner_discord_id` varchar(32) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_servers_discord` (`discord_guild_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "server_members" table
CREATE TABLE `server_members` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `joined_at` timestamp(6) NULL,
  `left_at` timestamp(6) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `idx_sm_server` (`server_id`),
  INDEX `idx_sm_user` (`user_id`),
  UNIQUE INDEX `uq_server_members` (`server_id`, `user_id`),
  CONSTRAINT `fk_sm_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_sm_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "udon_clients" table
CREATE TABLE `udon_clients` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `token_hash` char(64) NOT NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  `last_access_at` timestamp(6) NULL,
  `last_access_ip` varchar(45) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `idx_udon_clients_server` (`server_id`),
  CONSTRAINT `fk_udon_clients_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "vrc_registrations" table
CREATE TABLE `vrc_registrations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `vrc_name` varchar(64) NOT NULL,
  `vrc_name_norm` varchar(64) AS (lower(`vrc_name`)) STORED NULL,
  `status` enum('pending','approved','rejected') NOT NULL DEFAULT "pending",
  `reviewer_user_id` bigint unsigned NULL,
  `reviewed_at` timestamp(6) NULL,
  `note` varchar(255) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_vrc_reviewer` (`reviewer_user_id`),
  INDEX `fk_vrc_user` (`user_id`),
  INDEX `idx_vrc_by_user` (`server_id`, `user_id`, `status`),
  UNIQUE INDEX `uq_vrc_unique_active` (`server_id`, `vrc_name_norm`, `status`),
  CONSTRAINT `fk_vrc_reviewer` FOREIGN KEY (`reviewer_user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT `fk_vrc_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_vrc_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `chk_vrc_status` CHECK (`status` in (_utf8mb4'pending',_utf8mb4'approved',_utf8mb4'rejected'))
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "vrc_registration_events" table
CREATE TABLE `vrc_registration_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `registration_id` bigint unsigned NOT NULL,
  `event` enum('submitted','approved','rejected','revoked') NOT NULL,
  `actor_user_id` bigint unsigned NULL,
  `detail` varchar(255) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_vrc_ev_actor` (`actor_user_id`),
  INDEX `idx_vrc_ev_reg` (`registration_id`),
  CONSTRAINT `fk_vrc_ev_actor` FOREIGN KEY (`actor_user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT `fk_vrc_ev_reg` FOREIGN KEY (`registration_id`) REFERENCES `vrc_registrations` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "whitelists" table
CREATE TABLE `whitelists` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `version` bigint unsigned NOT NULL,
  `item_count` int unsigned NOT NULL DEFAULT 0,
  `path` varchar(255) NOT NULL,
  `content_hash` char(64) NOT NULL,
  `generated_by` bigint unsigned NULL,
  `generated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `expires_at` timestamp(6) NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  INDEX `fk_whitelists_generator` (`generated_by`),
  INDEX `idx_whitelists_active` (`server_id`, `is_active`),
  UNIQUE INDEX `uq_whitelists_ver` (`server_id`, `version`),
  CONSTRAINT `fk_whitelists_generator` FOREIGN KEY (`generated_by`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT `fk_whitelists_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "whitelist_items" table
CREATE TABLE `whitelist_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `whitelist_id` bigint unsigned NOT NULL,
  `vrc_name` varchar(64) NOT NULL,
  `vrc_name_norm` varchar(64) AS (lower(`vrc_name`)) STORED NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_wli_norm` (`vrc_name_norm`),
  UNIQUE INDEX `uq_whitelist_item` (`whitelist_id`, `vrc_name_norm`),
  CONSTRAINT `fk_wli_whitelist` FOREIGN KEY (`whitelist_id`) REFERENCES `whitelists` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
