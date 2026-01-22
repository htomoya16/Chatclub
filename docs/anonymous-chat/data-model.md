# 匿名チャット機能 データモデル

本機能は**投稿内容を保存しない**。
永続化が必要なのは「匿名チャンネルの設定」のみ。

---

## 1. anonymous_channels

| カラム | 型 | 説明 |
| --- | --- | --- |
| guild_id | text | Discord Guild ID |
| channel_id | text | Discord Channel ID |
| webhook_id | text | Webhook ID |
| webhook_token | text | Webhook Token（秘匿情報） |
| created_at | timestamptz | 作成日時（UTC） |
| updated_at | timestamptz | 更新日時（UTC） |

### 制約

- `primary key (guild_id, channel_id)`
- `foreign key (guild_id) references guilds (id) on delete cascade`

---

## 2. 非保存データ

- 投稿本文
- 添付ファイル
- 投稿者ID

---

## 3. 補足

- Discord ID は文字列として扱う
- Webhook Token は漏洩を防ぐため保護対象とする
