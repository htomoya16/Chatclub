# SF6 Buckler Commands

本ドキュメントは、Buckler 対戦ログ機能の Discord コマンド仕様を定義する。

現状:

- 実装済み: `/sf6_account`, `/sf6_unlink`, `/sf6_fetch`, `/sf6_friend`, `/sf6_stats`, `/sf6_session`, `/sf6_history`

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
- 出力: 解除完了メッセージ（戦績は保持）

---

## 2. 友達（対戦相手）管理

### /sf6_friend

- 概要: 対戦相手一覧と操作ボタンを表示
- 入力: なし（ボタン押下で追加/削除）
- 出力: 友達一覧 Embed

---

## 3. セッション統計

### /sf6_session start

- 概要: 対戦セッションを開始（ユーザーごとに1つだけ、再開は上書き）
- 入力:
  - opponent_code (sid) 必須
  - subject_code (sid) 任意（未指定なら連携アカウント）
- 出力: セッション開始メッセージ

### /sf6_session end

- 概要: 対戦セッションを終了し、期間内の統計を表示
- 入力:
  - opponent_code (sid) 必須
  - subject_code (sid) 任意（未指定なら連携アカウント）
- 出力: セッション統計（勝率・勝敗・キャラ別）

---

## 4. 統計表示（都度クエリ）

### /sf6_stats range

- 概要: 期間指定の統計を表示する（JST）
- 入力:
  - opponent_code (sid) 必須
  - subject_code (sid) 任意（未指定なら連携アカウント）
  - from (YYYY-MM-DD, JST) 必須
  - to (YYYY-MM-DD, JST) 必須
- 出力: 総試合数 / 勝・敗・引き分け / 勝率（分除外）/ キャラ別勝率

### /sf6_stats count

- 概要: 直近 N 戦の統計を表示する
- 入力:
  - opponent_code (sid) 必須
  - subject_code (sid) 任意（未指定なら連携アカウント）
  - count (N) 必須
- 出力: 総試合数 / 勝・敗・引き分け / 勝率（分除外）/ キャラ別勝率

---

### /sf6_stats set

- 概要: 直近の対戦を **30分以内の間隔でグルーピング**し、まとめた統計を表示する
- 入力:
  - opponent_code (sid) 必須
  - subject_code (sid) 任意（未指定なら連携アカウント）
- 出力:
  - 期間（JST）/ 合計試合数 / 勝敗 / 勝率 / キャラ別勝率
  - 1グループ=1ページ（前へ/次へで過去分）

---

## 5. 履歴表示

### /sf6_history

- 概要: 対戦履歴（Custom）の一覧を表示する
- 入力:
  - opponent_code (sid) 必須
  - subject_code (sid) 任意（未指定なら連携アカウント）
- 出力:
  - 日時（JST）/ 左に subject（連携済みならメンション）+ 使用キャラ
  - 右に opponent（連携済みならメンション）+ 使用キャラ
  - 勝敗（例: LOSE vs WIN）
- 備考:
  - 5件/ページでページング（前へ/次へボタン）

---

## 6. 取得（現行）

### /sf6_fetch

- 概要: ギルド内の登録済みアカウント/フレンドの Battle Log（Custom）を取得して保存する（管理者/許可ユーザー限定）
- 入力: なし
- 挙動:
  - アカウントの sid を優先して取得
  - 同一 sid がフレンド登録されている場合はフレンド側をスキップ
  - 最大 10 ページまで取得（既存データで早期終了）
- 出力: 保存件数 / 取得ページ数 / スキップ数
