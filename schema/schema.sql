-- Create "users" table
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `discord_id` varchar(32) NOT NULL,
  `name` varchar(191) NULL,
  `avatar_url` varchar(255) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_users_discord` (`discord_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "beats" table
CREATE TABLE `beats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NOT NULL,
  `youtube_url` varchar(500) NOT NULL,
  `video_id` varchar(32) NOT NULL,
  `is_enabled` bool NOT NULL DEFAULT 1,
  `created_by` bigint unsigned NOT NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `idx_beats_created_by` (`created_by`),
  INDEX `idx_beats_is_enabled` (`is_enabled`),
  UNIQUE INDEX `uq_beats_video_id` (`video_id`),
  CONSTRAINT `fk_beats_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
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
-- Create "cypher_sessions" table
CREATE TABLE `cypher_sessions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `channel_id` varchar(32) NOT NULL,
  `message_id` varchar(32) NULL,
  `beat_id` bigint unsigned NOT NULL,
  `started_by` bigint unsigned NOT NULL,
  `status` enum('playing','stopped') NOT NULL DEFAULT "playing",
  `started_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `ended_at` timestamp(6) NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_cypher_beat` (`beat_id`),
  INDEX `fk_cypher_starter` (`started_by`),
  INDEX `idx_cypher_server` (`server_id`),
  INDEX `idx_cypher_status` (`status`),
  CONSTRAINT `fk_cypher_beat` FOREIGN KEY (`beat_id`) REFERENCES `beats` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_cypher_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_cypher_starter` FOREIGN KEY (`started_by`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
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
-- Create "tournaments" table
CREATE TABLE `tournaments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `thread_id` varchar(32) NOT NULL,
  `name` varchar(191) NOT NULL,
  `max_players` int unsigned NOT NULL,
  `status` enum('recruiting','ongoing','finished','canceled') NOT NULL DEFAULT "recruiting",
  `created_by` bigint unsigned NOT NULL,
  `started_at` datetime(6) NULL,
  `finished_at` datetime(6) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_tournaments_creator` (`created_by`),
  INDEX `idx_tournaments_server` (`server_id`),
  INDEX `idx_tournaments_thread` (`thread_id`),
  CONSTRAINT `fk_tournaments_creator` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_tournaments_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "tournament_entries" table
CREATE TABLE `tournament_entries` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tournament_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  `joined_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `left_at` timestamp(6) NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_tournament_entries_tournament` (`tournament_id`),
  INDEX `idx_tournament_entries_user` (`user_id`),
  UNIQUE INDEX `uq_tournament_entry_active` (`tournament_id`, `user_id`, `is_active`),
  CONSTRAINT `fk_tournament_entries_tournament` FOREIGN KEY (`tournament_id`) REFERENCES `tournaments` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_tournament_entries_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "tournament_matches" table
CREATE TABLE `tournament_matches` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tournament_id` bigint unsigned NOT NULL,
  `round_no` int unsigned NOT NULL,
  `match_no` int unsigned NOT NULL,
  `player1_entry_id` bigint unsigned NOT NULL,
  `player2_entry_id` bigint unsigned NOT NULL,
  `beat_id` bigint unsigned NULL,
  `status` enum('pending','ongoing','voting','finished','canceled') NOT NULL DEFAULT "pending",
  `winner_entry_id` bigint unsigned NULL,
  `started_at` datetime(6) NULL,
  `voting_started_at` datetime(6) NULL,
  `finished_at` datetime(6) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_tm_beat` (`beat_id`),
  INDEX `fk_tm_p1` (`player1_entry_id`),
  INDEX `fk_tm_p2` (`player2_entry_id`),
  INDEX `fk_tm_winner` (`winner_entry_id`),
  INDEX `idx_matches_status` (`status`),
  INDEX `idx_matches_tournament` (`tournament_id`),
  UNIQUE INDEX `uq_match_in_tournament` (`tournament_id`, `round_no`, `match_no`),
  CONSTRAINT `fk_tm_beat` FOREIGN KEY (`beat_id`) REFERENCES `beats` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT `fk_tm_p1` FOREIGN KEY (`player1_entry_id`) REFERENCES `tournament_entries` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_tm_p2` FOREIGN KEY (`player2_entry_id`) REFERENCES `tournament_entries` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_tm_tournament` FOREIGN KEY (`tournament_id`) REFERENCES `tournaments` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_tm_winner` FOREIGN KEY (`winner_entry_id`) REFERENCES `tournament_entries` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT `chk_tm_distinct_players` CHECK (`player1_entry_id` <> `player2_entry_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "tournament_votes" table
CREATE TABLE `tournament_votes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `match_id` bigint unsigned NOT NULL,
  `voter_user_id` bigint unsigned NOT NULL,
  `voted_entry_id` bigint unsigned NOT NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_votes_voter` (`voter_user_id`),
  INDEX `idx_votes_match` (`match_id`),
  INDEX `idx_votes_voted_entry` (`voted_entry_id`),
  UNIQUE INDEX `uq_vote_unique` (`match_id`, `voter_user_id`),
  CONSTRAINT `fk_votes_match` FOREIGN KEY (`match_id`) REFERENCES `tournament_matches` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_votes_voted_entry` FOREIGN KEY (`voted_entry_id`) REFERENCES `tournament_entries` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_votes_voter` FOREIGN KEY (`voter_user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
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
CREATE TABLE whitelists (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  platform ENUM('discord', 'vrc') NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  note VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_platform_user (platform, user_id)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
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
