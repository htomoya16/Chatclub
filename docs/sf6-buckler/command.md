# SF6 Buckler Commands

本ドキュメントは、Buckler 対戦ログ機能の Discord コマンド仕様を定義する。

現状:

- 実装済み: `/sf6_account`, `/sf6_unlink`, `/sf6_fetch`
- 未実装: `/sf6_friend *`, `/watch_*`, `/sf6_stats`（設計のみ）

補足: 本ドキュメントの fighter_id は **Buckler プロフィールの short_id（sid（ユーザーコード））** を指す。

---

## 1. アカウント連携

### /sf6_account

- 概要: **リンク状態パネルを表示**し、ボタンから登録/解除できる
- 入力: なし（ボタン押下で user_code を入力）
- 出力: 状態表示 Embed（未連携=赤 / 連携済み=緑）
- 備考: キャラ画像表示のため、本番運用では `PUBLIC_BASE_URL` に **外部公開URL** を設定する

### /sf6_unlink

- 概要: 連携解除（収集停止）
- 出力: 解除完了メッセージ（戦績も削除）

---

## 2. 友達（対戦相手）管理

### /sf6_friend add

- 概要: 対戦相手を登録する
- 入力: fighter_id（sid（ユーザーコード））, alias (optional)
- 出力: 登録完了メッセージ

### /sf6_friend remove

- 概要: 対戦相手を削除する
- 入力: fighter_id（sid（ユーザーコード）） または alias
- 出力: 削除完了メッセージ

### /sf6_friend list

- 概要: 登録済み友達を一覧する
- 出力: fighter_id / alias の一覧

---

## 3. セッション監視

### /watch_start

- 概要: 指定した友達とのセッション監視を開始する
- 入力: opponent（fighter_id / sid（ユーザーコード） または alias）
- 出力: 監視開始メッセージ

### /watch_end

- 概要: 監視中セッションを終了する
- 入力: なし
- 出力: セッション結果サマリ

---

## 4. 統計表示

### /sf6_stats

- 概要: 戦績統計を表示する
- 入力: opponent (optional), recent_n (optional), period (optional)
- 出力: 勝率 / 試合数 / 直近 N 戦 / 連勝・連敗 / キャラ別

---

## 5. 取得（現行）

### /sf6_fetch

- 概要: 指定した user_code の Battle Log（Custom）を取得して保存する
- 入力: user_code（sid（ユーザーコード））, page (optional, default=1)
- 挙動:
  - 最大 10 ページまで取得
  - `source_key` が全件既存なら早期終了
  - 取得は Buckler の data API を使用
- 出力: 保存件数 / 取得ページ数
