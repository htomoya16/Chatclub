# SF6 Buckler Data Model

本ドキュメントは、SF6 Buckler 対戦ログ機能で必要となる
最小限のテーブル構成と制約を定義する。

---

## 1. テーブル一覧

### 1.1 sf6_accounts

Discord ユーザと CFN アカウントの紐づけ。

| column | type | description |
| --- | --- | --- |
| id | uuid | primary key |
| guild_id | text | Discord Guild ID |
| user_id | text | Discord User ID |
| fighter_id | text | Buckler の short_id（sid） |
| display_name | text | Buckler の fighter_id（表示名, 任意） |
| status | text | active / inactive |
| created_at | timestamptz | created time (UTC) |
| updated_at | timestamptz | updated time (UTC) |

Constraints:

- unique (guild_id, user_id)
- unique (guild_id, fighter_id)

---

### 1.2 sf6_friends

対戦相手（友達）を保持する。

| column | type | description |
| --- | --- | --- |
| id | uuid | primary key |
| guild_id | text | Discord Guild ID |
| user_id | text | Discord User ID（登録者） |
| fighter_id | text | Buckler の short_id（sid） |
| display_name | text | Buckler の fighter_id（表示名, 任意） |
| alias | text | 別名（任意） |
| created_at | timestamptz | created time (UTC) |
| updated_at | timestamptz | updated time (UTC) |

Constraints:

- unique (guild_id, user_id, fighter_id)

---

### 1.3 sf6_sessions

セッション監視の状態を保持する。

| column | type | description |
| --- | --- | --- |
| id | uuid | primary key |
| guild_id | text | Discord Guild ID |
| user_id | text | Discord User ID |
| opponent_fighter_id | text | 友達の short_id（sid） |
| status | text | active / ended |
| started_at | timestamptz | セッション開始時刻 (UTC) |
| ended_at | timestamptz | セッション終了時刻 (UTC, nullable) |
| last_polled_at | timestamptz | 最終ポーリング時刻 (UTC) |
| last_seen_battle_at | timestamptz | 最新試合の時刻 (UTC, nullable) |
| created_at | timestamptz | created time (UTC) |
| updated_at | timestamptz | updated time (UTC) |

Constraints:

- index (guild_id, user_id, status)
- index (guild_id, user_id, opponent_fighter_id)

---

### 1.4 sf6_battles

取得したカスタムマッチ戦績を保存する。

| column | type | description |
| --- | --- | --- |
| id | uuid | primary key |
| guild_id | text | Discord Guild ID |
| user_id | text | Discord User ID |
| opponent_fighter_id | text | 友達の short_id（sid） |
| battle_at | timestamptz | 試合時刻 (UTC) |
| result | text | win / loss / draw |
| self_character | text | 自キャラ |
| opponent_character | text | 相手キャラ |
| round_wins | int | 自分のラウンド数（round_results の >0 数） |
| round_losses | int | 相手のラウンド数（round_results の >0 数） |
| source_battle_id | text | Buckler 側の試合 ID（replay_id） |
| source_key | text | 重複排除用のユニークキー |
| session_id | uuid | sf6_sessions.id (nullable) |
| raw_payload | jsonb | 取得データの原文 (任意) |
| created_at | timestamptz | created time (UTC) |
| updated_at | timestamptz | updated time (UTC) |

Constraints:

- unique (guild_id, user_id, source_key)
- index (guild_id, user_id, opponent_fighter_id, battle_at)
- index (guild_id, user_id, battle_at)

---

## 2. 重複排除の考え方

- Buckler の Battle Log に **replay_id** が存在するため `source_battle_id` に利用する
- 試合 ID が無い場合は以下を結合して `source_key` を生成する
  - battle_at（秒精度）
  - 自分 / 相手の short_id（sid）
  - 自キャラ / 相手キャラ
  - 勝敗
- `source_key` は **安定した文字列**で生成し、ユニーク制約で重複排除する

---

## 3. 注意点

- DB 保存は **UTC 固定**
- Battle Log の構造は変更される可能性があるため `raw_payload` を保持可能にする
- セッション関連の集計は `sf6_sessions` に紐づけて行う
- 本ドキュメントの fighter_id は **Buckler の short_id（sid）** を指す
