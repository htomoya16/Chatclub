# SF6 Buckler Flow

本ドキュメントは、Buckler 対戦ログ機能の主要な処理フローを整理する。

---

## 0. 認証/セッション管理フロー

1. 自動ログインで Buckler の認証 Cookie を取得する
2. Cookie を安全に保管し、取得リクエストに付与する
3. 403/ログイン要求などの失効兆候が出たら再ログインする
4. 再ログイン失敗時は収集を停止し、管理者に通知する

### 0.1 HTTPログインフロー（直叩き）

1. `GET https://auth.cid.capcom.com/login?...` でログインページを取得
2. `POST /usernamepassword/challenge`（JSON: `state`）
3. `POST /usernamepassword/login`（JSON: `client_id`, `connection`, `username`, `password`, `redirect_uri`, `state`, `_csrf` など）
4. レスポンスHTML内の `/login/callback` フォームを `POST`（hidden input を送信）
5. `GET https://cid.capcom.com/ja/loginCallback?code=***&state=***`
6. `GET https://auth.cid.capcom.com/authorize?...`
7. `GET https://www.streetfighter.com/6/buckler/auth/login?code=***&state=***`
   - ここで `buckler_id` / `buckler_r_id` が発行される
8. `GET /6/buckler/ja-jp/auth/postlogin` → `/6/buckler/ja-jp/?status=login`

### 0.2 Cookie運用

- 必須Cookie: `buckler_id` / `buckler_r_id`
- Cookieは暗号化して保存し、失効時は再ログインで更新する
- 2FA/CAPTCHAなしを前提（有効化された場合は自動ログイン不可）

---

## 1. 友達登録フロー

1. ユーザが `/sf6_friend add` を実行
2. fighter_id を登録し、任意で alias を保存
3. 同一ユーザ + fighter_id は重複登録を禁止
4. 登録完了メッセージを返す

---

## 2. 定期ポーリングフロー

1. スケジューラが **一定間隔**で Buckler Battle Log を取得
2. buildId をキャッシュから取得し、無ければ HTML から抽出
3. buildId で data API を呼び出す
   - `/6/buckler/_next/data/{buildId}/ja-jp/profile/{sid}/battlelog/custom.json?sid={sid}&page={page}`
   - ページサイズは 10 件、`page` クエリでページング
4. 404/410 が返ったら buildId を再取得して再試行
5. 直近 N 件になるまでページを進めて `replay_list` を収集
6. Custom Room の試合のみ抽出（`replay_battle_type == 4`）
7. 「自分 vs 登録友達」に一致する試合だけを残す
8. `source_key` による重複排除（UPSERT）
9. 新規試合があれば統計を更新し、必要なら通知する

補足:

- 取得間隔は Buckler の負荷と Discord 通知の要件で調整する
- 連続失敗時は指数バックオフする
- 自分/相手の判定は `pageProps.sid` と `player*.player.short_id` を突き合わせる
- buildId は一定期間ごとに再取得してもよい（例: 1日1回）

---

## 3. セッション監視フロー

### 3.1 監視開始

1. `/watch_start opponent=<友達>` で開始
2. `sf6_sessions` に active セッションを作成
3. 以降のポーリング間隔を短縮（例: 30〜90 秒）

### 3.2 監視中の増分取得

1. Battle Log を取得
2. セッション開始時刻以降の試合に限定
3. Custom Room かつ「自分 vs 友達」に一致する試合だけ保存
4. `round_results` から勝敗/ラウンド数を算出
5. 新規試合があれば **セッション統計**を更新

### 3.3 監視終了

- `/watch_end` で手動終了
- もしくは **一定時間更新なし**で自動終了（例: 20〜30 分）
- 終了時に直近 N 件を再取得し取りこぼし補完
- セッション結果を Discord に要約表示

---

## 4. 取得漏れ対策

- **直近 N 件の再取得 + 重複排除**を標準動作とする
- セッション終了時に再取得を行い、取りこぼしを埋める

---

## 5. 失敗時の扱い

- Buckler 側エラー時は既存データを維持する
- 一定回数失敗で監視終了（運用判断）
- 例外は Bot 全体を停止させない（要隔離）
