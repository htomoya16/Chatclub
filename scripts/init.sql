-- データベース作成
CREATE DATABASE IF NOT EXISTS yasairap
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_0900_ai_ci;

USE yasairap;

-- Discordユーザ
CREATE TABLE IF NOT EXISTS users (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  discord_id    VARCHAR(32) NOT NULL,                 -- 雪だるま式の桁に備え可変長
  name          VARCHAR(191) NULL,
  avatar_url    VARCHAR(255) NULL,
  created_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_users_discord (discord_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- チーム（個人戦のみなら空テーブルで問題ない）
CREATE TABLE IF NOT EXISTS teams (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name          VARCHAR(191) NOT NULL,
  owner_user_id BIGINT UNSIGNED NULL,
  created_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_teams_name (name),
  CONSTRAINT fk_teams_owner
    FOREIGN KEY (owner_user_id) REFERENCES users(id)
    ON UPDATE RESTRICT ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- シーズン（大会期）
CREATE TABLE IF NOT EXISTS seasons (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  slug          VARCHAR(64) NOT NULL,                 -- 例: 2025-fall
  title         VARCHAR(191) NOT NULL,
  ruleset       ENUM('league','single_elim','double_elim') NOT NULL DEFAULT 'league',
  starts_on     DATE NULL,
  ends_on       DATE NULL,
  status        ENUM('draft','open','running','closed') NOT NULL DEFAULT 'draft',
  created_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_seasons_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 参加者（個人またはチームのどちらか一方）
CREATE TABLE IF NOT EXISTS entries (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  season_id     BIGINT UNSIGNED NOT NULL,
  kind          ENUM('solo','team') NOT NULL,
  user_id       BIGINT UNSIGNED NULL,
  team_id       BIGINT UNSIGNED NULL,
  display_name  VARCHAR(191) NULL,                    -- 表示名（固定化したい場合に使う）
  created_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_entries_unique (season_id, kind, user_id, team_id),
  KEY idx_entries_season (season_id),
  CONSTRAINT fk_entries_season
    FOREIGN KEY (season_id) REFERENCES seasons(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT fk_entries_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT fk_entries_team
    FOREIGN KEY (team_id) REFERENCES teams(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  -- 個人戦orチーム戦の排他（MySQL 8 は CHECK を評価する）
  CONSTRAINT chk_entries_exact_one CHECK (
    (kind = 'solo' AND user_id IS NOT NULL AND team_id IS NULL)
    OR
    (kind = 'team' AND team_id IS NOT NULL AND user_id IS NULL)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 試合
CREATE TABLE IF NOT EXISTS matches (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  season_id       BIGINT UNSIGNED NOT NULL,
  round_no        INT UNSIGNED NULL,                  -- リーグの節やトナメのラウンド
  bracket         ENUM('main','losers') NULL,         -- ダブルエリミ用
  home_entry_id   BIGINT UNSIGNED NOT NULL,
  away_entry_id   BIGINT UNSIGNED NOT NULL,
  scheduled_at    DATETIME(6) NULL,
  status          ENUM('pending','scheduled','completed','canceled')
                  NOT NULL DEFAULT 'pending',
  created_at      TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at      TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_matches_season_round (season_id, round_no),
  KEY idx_matches_time (season_id, scheduled_at),
  CONSTRAINT fk_matches_season
    FOREIGN KEY (season_id) REFERENCES seasons(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT fk_matches_home
    FOREIGN KEY (home_entry_id) REFERENCES entries(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT fk_matches_away
    FOREIGN KEY (away_entry_id) REFERENCES entries(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT chk_matches_distinct CHECK (home_entry_id <> away_entry_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 結果（1試合1レコード）
CREATE TABLE IF NOT EXISTS results (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  match_id        BIGINT UNSIGNED NOT NULL,
  score_home      INT NOT NULL,
  score_away      INT NOT NULL,
  method          ENUM('normal','ff','dq') NOT NULL DEFAULT 'normal', -- 不戦/失格など
  reported_by     BIGINT UNSIGNED NULL,            -- users.id（報告者）
  reported_at     TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  note            VARCHAR(255) NULL,
  created_at      TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at      TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_results_match (match_id),
  CONSTRAINT fk_results_match
    FOREIGN KEY (match_id) REFERENCES matches(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT fk_results_reporter
    FOREIGN KEY (reported_by) REFERENCES users(id)
    ON UPDATE RESTRICT ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- レーティング（任意：スナップショット方式）
CREATE TABLE IF NOT EXISTS ratings (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  season_id       BIGINT UNSIGNED NOT NULL,
  entry_id        BIGINT UNSIGNED NOT NULL,
  rating          DOUBLE NOT NULL,
  rd              DOUBLE NULL,                        -- 変動幅（Glicko等を想定）
  computed_at     TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uq_ratings_point (season_id, entry_id, computed_at),
  KEY idx_ratings_latest (season_id, entry_id, computed_at),
  CONSTRAINT fk_ratings_season
    FOREIGN KEY (season_id) REFERENCES seasons(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT fk_ratings_entry
    FOREIGN KEY (entry_id) REFERENCES entries(id)
    ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 生成アセット（PNGなど）
CREATE TABLE IF NOT EXISTS assets (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  kind            ENUM('standings','schedule','profile') NOT NULL,
  season_id       BIGINT UNSIGNED NULL,
  entry_id        BIGINT UNSIGNED NULL,               -- profile の場合など
  path            VARCHAR(255) NOT NULL,              -- 例: s3://... or https://...
  hash            CHAR(40) NULL,                      -- 内容ハッシュ（rev用）
  updated_at      TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  created_at      TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_assets_season (season_id, kind),
  KEY idx_assets_entry (entry_id, kind),
  CONSTRAINT fk_assets_season
    FOREIGN KEY (season_id) REFERENCES seasons(id)
    ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT fk_assets_entry
    FOREIGN KEY (entry_id) REFERENCES entries(id)
    ON UPDATE RESTRICT ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
