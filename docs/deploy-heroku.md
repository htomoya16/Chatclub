# Heroku Deploy

このドキュメントは、Chatclub を既存または新規の Heroku app にデプロイする手順をまとめる。

Heroku では `docker-compose.yml` の PostgreSQL コンテナは使わない。DB は Heroku Postgres add-on を使い、アプリは Heroku が提供する `DATABASE_URL` に接続する。

---

## 1. 前提

- Heroku CLI にログイン済み
- Heroku app を作成済み、または差し替え先の app 名が決まっている
- Discord Developer Portal の Bot token / Application ID を取得済み
- SF6 Buckler 機能を使う場合は CAPCOM / Buckler 関連の環境変数を用意済み

---

## 2. Heroku app を準備

新規 app の場合:

```bash
heroku create <APP_NAME>
```

既存 app を差し替える場合:

```bash
heroku git:remote -a <APP_NAME>
```

Dockerfile からビルドするため、stack を `container` にする。
Heroku app を Dockerfile / heroku.yml でデプロイする設定に切り替えるコマンド

```bash
heroku stack:set container -a <APP_NAME>
```

---

## 3. Heroku Postgres を追加

```bash
heroku addons:create heroku-postgresql:essential-0 -a <APP_NAME>
```

Heroku Postgres を追加すると、`DATABASE_URL` が config var として自動設定される。

---

## 4. Config Vars を設定

必須:

```bash
heroku config:set \
  DISCORD_TOKEN="<discord bot token>" \
  DISCORD_APP_ID="<discord app id>" \
  DISCORD_GUILD_IDS="<guild id>" \
  DISCORD_REGISTER_COMMANDS=true \
  DB_SSLMODE=require \
  PUBLIC_BASE_URL="https://<APP_NAME>.herokuapp.com" \
  -a <APP_NAME>
```

SF6 Buckler 機能を使う場合:

```bash
heroku config:set \
  CAPCOM_EMAIL="<capcom email>" \
  CAPCOM_PASSWORD="<capcom password>" \
  BUCKLER_COOKIE_ENC_KEY="<32 byte raw/base64/hex key>" \
  BUCKLER_CLIENT_ID="<buckler client id>" \
  BUCKLER_LANG=ja-jp \
  BUCKLER_BASE_URL="https://www.streetfighter.com/6/buckler" \
  -a <APP_NAME>
```

`BUCKLER_COOKIE_ENC_KEY` は以下で生成できる。

```bash
openssl rand -hex 32
```

`BUCKLER_COOKIE_ENC_KEY` は Cookie の値ではなく、Cookie を暗号化保存するための鍵である。
Buckler の Cookie は `CAPCOM_EMAIL` / `CAPCOM_PASSWORD` による自動ログインで取得する。
詳細は `docs/sf6-buckler/flow.md` の「Cookie運用」を参照する。

---

## 5. デプロイ

```bash
git push heroku main
```

ブランチ名が `main` ではない場合:

```bash
git push heroku HEAD:main
```

---

## 6. マイグレーション

Atlas をローカルにインストールしていない場合:

```bash
curl -sSf https://atlasgo.sh | sh
```

Heroku Postgres にマイグレーションを適用する。

```bash
atlas migrate apply \
  --dir "file://migrations" \
  --url "$(heroku config:get DATABASE_URL -a <APP_NAME>)?sslmode=require"
```

---

## 7. 動作確認

```bash
curl https://<APP_NAME>.herokuapp.com/api/healthz
heroku logs --tail -a <APP_NAME>
```

Discord サーバで `/ping` が返れば Bot 側も動作している。

コマンド登録が終わったら、毎回登録しないように戻す。

```bash
heroku config:set DISCORD_REGISTER_COMMANDS=false -a <APP_NAME>
```

Discord Bot を常時オンラインにしたい場合、Eco dyno は無通信時に sleep するため避ける。

```bash
heroku ps:scale web=1:basic -a <APP_NAME>
```

---

## 8. 注意

- Heroku Postgres は SSL 接続が必要なため、`DATABASE_URL` 使用時はアプリ側で `sslmode=require` を補完する。
- Heroku では `PORT` は自動設定される。手動で固定しない。
- Web dyno を 2 台以上にすると Discord Gateway 接続も複数立つため、`web=1` で運用する。
- Eco dyno は sleep するため、Discord Bot の常時稼働には Basic 以上を使う。
- 既存 app を差し替える場合、既存 DB を残すか新規 Heroku Postgres にするかを先に決める。
