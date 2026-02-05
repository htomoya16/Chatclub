# Setup

本ドキュメントは、開発・運用のためのセットアップ手順をまとめる。

---

## 1. リポジトリ取得

```bash
git clone <YOUR_REPO_URL>
cd Chatclub
```

---

## 2. 環境変数の設定

`.env.example` から `.env` を作成し、必要な値を設定する。

```env
# PostgreSQL
POSTGRES_USER=chatclub_user
POSTGRES_PASSWORD=changeme
POSTGRES_DB=chatclub
POSTGRES_PORT=5432
POSTGRES_TZ=UTC

# アプリ
APP_PORT=8080
DB_HOST=postgres
DB_PORT=5432

# DISCORD関連
DISCORD_TOKEN=your_discord_bot_token
DISCORD_APP_ID=your_discord_app_id
DISCORD_GUILD_IDS=your_test_guild_id # カンマ区切り。空ならグローバルコマンド（反映に時間がかかる）
```

---

## 3. プロジェクトを起動（開発環境）

### 初回

```bash
docker compose --profile dev up --build
```

### バックグラウンド起動

```bash
docker compose --profile dev up -d --build
```

### 停止

```bash
docker compose --profile dev down
```

---

## 4. プロジェクトを起動（本番環境）

### 初回

```bash
docker compose --profile prod up --build
```

### バックグラウンド起動

```bash
docker compose --profile prod up -d --build
```

### 停止

```bash
docker compose --profile prod down
```

---

## 5. Atlas によるマイグレーション適用

### Atlas のインストール（WSL 上で実行）

```bash
curl -sSf https://atlasgo.sh | sh
```

### マイグレーション適用（docker compose up 済みの状態）

```bash
atlas migrate apply --env local
```

---

## 6. Discord Bot セットアップ

このバックエンドは Discord Bot を通じて操作される。  
以下の手順で Discord Developer Portal 上に Bot を作成し、環境変数に必要な値を設定する。

### 6.1 アプリケーションの作成

1. [Discord Developer Portal](https://discord.com/developers/applications) にアクセスし、ログインする。  
2. 「**New Application**」をクリックして新しいアプリケーションを作成。  
3. 作成後、左メニューから **Bot** を選び、「Add Bot」→「Yes, do it!」をクリック。  
4. 作成された Bot のトークンを `.env` に設定する。

```env
DISCORD_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 6.2 Application ID と Guild ID の取得

#### (1) Application ID

1. Developer Portal のアプリケーションページで対象アプリを選択し **General Information** を開く。  
2. 「Application ID」をコピーして `.env` の `DISCORD_APP_ID` に設定する。

#### (2) Guild ID（サーバID）

1. Discord クライアントの「設定 → 詳細設定」から **開発者モード** を有効にする。  
2. Bot をテストする Discord サーバで、サーバアイコンを右クリック → 「IDをコピー」。  
3. `.env` の `DISCORD_GUILD_IDS` に貼り付ける（複数ある場合はカンマ区切り）。  
   グローバルコマンドにしたい場合は `DISCORD_GUILD_IDS` を空にする（反映に時間がかかる）。

### 6.3 Bot をサーバに招待

1. Developer Portal の **OAuth2 → URL Generator** を開く。  
2. 「**bot**」と「**applications.commands**」にチェックを入れる。  
3. 「Bot Permissions」で以下を選択：
   - Send Messages  
   - Read Message History
   - View Channels  
   - Manage Messages  
   - Use Slash Commands  
4. 生成された URL をブラウザで開き、テスト用サーバに Bot を追加する。

### 6.4 Intent の設定

Bot がメッセージ内容やメンバー情報にアクセスできるようにするため、  
**Bot → Privileged Gateway Intents** で以下を有効化しておく。

- ✅ **MESSAGE CONTENT INTENT**  
- ✅ **SERVER MEMBERS INTENT**

### 6.5 動作確認

環境変数が設定された状態で `docker compose up --build` し、  
Discord サーバで Bot がオンラインになれば成功。
