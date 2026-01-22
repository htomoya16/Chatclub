# Chatclub Backend

**Chatclub Backend** は、Discord サーバ内の**通話活動（VC）**や**ゲーム活動**を記録・集計し、  
Discord Bot と Web から可視化するためのバックエンドシステムである。

---

## ⚙️ 技術スタック

| 項目 | 使用技術 |
|------|-----------|
| 言語 | Go 1.25.1 |
| Web Framework | Echo v4 |
| DB | PostgreSQL 16 |
| Migration | Atlas |
| Container | Docker / docker-compose |
| Discord連携 | [discordgo](https://github.com/bwmarrin/discordgo) |


## 🐳 Dockerセットアップ

### 1. リポジトリ取得
```bash
git clone <YOUR_REPO_URL>
cd Chatclub
```

### 2. 環境変数の設定

`.env.example`より`.env` ファイルを作成し、以下を設定する。

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
DISCORD_GUILD_ID=your_test_guild_id
```

### 3. プロジェクトを起動(開発環境)
#### 初回
```bash
# Dockerコンテナを起動
docker compose up --build

# バックグラウンドで起動する場合
docker compose up -d --build
```

#### 初回以降
```bash
# Dockerコンテナを起動
docker compose --profile dev up

# バックグラウンドで起動する場合
docker compose --profile dev up -d
```

#### 止め方
```bash
docker compose --profile dev down
```

### 4. プロジェクトを起動(本番環境)
#### 初回
```bash
# Dockerコンテナを起動
docker compose --profile prod up --build

# バックグラウンドで起動する場合
docker compose --profile prod up -d --build
```

#### 初回以降
```bash
# Dockerコンテナを起動
docker compose --profile prod up

# バックグラウンドで起動する場合
docker compose --profile prod up -d
```

#### 止め方
```bash
docker compose --profile prod down
```

### 5. Atlas によるマイグレーション適用
#### Atlas のインストール（WSL 上で実行）
```bash
curl -sSf https://atlasgo.sh | sh
```

####　マイグレーション適用(4. でdocker compose upした状態)
```bash
atlas migrate apply --env local
```

## 🤖 Discord Bot セットアップ

このバックエンドは Discord Bot を通じて操作される。  
以下の手順で Discord Developer Portal 上に Bot を作成し、環境変数に必要な値を設定する。

### 1. アプリケーションの作成

1. [Discord Developer Portal](https://discord.com/developers/applications) にアクセスし、ログインする。  
2. 「**New Application**」をクリックして新しいアプリケーションを作成。  
   名前は例として `Chatclub Bot` にしておくと分かりやすい。  
3. 作成後、左メニューから **Bot** を選び、「Add Bot」→「Yes, do it!」をクリック。  
4. 作成された Bot のトークンをコピーして `.env` に設定する(Dockerセットアップ2.)。
```env
DISCORD_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 2. Application ID と Guild ID の取得

#### (1) Application ID

1. Developer Portal のアプリケーションページで`Chatclub Bot`を選択し **General Information** を開く。  
2. 「Application ID」をコピーして `.env` の `DISCORD_APP_ID` に設定する。

#### (2) Guild ID（サーバID）

1. Discord クライアントの「設定 → 詳細設定」から **開発者モード** を有効にする。  
2. Bot をテストする Discord サーバで、サーバアイコンを右クリック → 「IDをコピー」。  
3. `.env` の `DISCORD_GUILD_ID` に貼り付ける。

### 3. Bot をサーバに招待

1. Developer Portal の **OAuth2 → URL Generator** を開く。  
2. 「**bot**」と「**applications.commands**」にチェックを入れる。  
3. 「Bot Permissions」で以下を選択：
   - Send Messages  
   - Read Message History
   - View Channels  
   - Manage Messages  
   - Use Slash Commands  
4. 生成された URL をブラウザで開き、テスト用サーバに Bot を追加する。

### 4. Intent の設定

Bot がメッセージ内容やメンバー情報にアクセスできるようにするため、  
**Bot → Privileged Gateway Intents** で以下を有効化しておく。

- ✅ **MESSAGE CONTENT INTENT**  
- ✅ **SERVER MEMBERS INTENT**

### 5. 動作確認

環境変数が設定された状態でアプリを起動(Dockerセットアップ 3.で```docker compose up --build```する)後、Discord サーバで Bot がオンラインになれば成功。
