# SF6 Buckler Commands

本ドキュメントは、Buckler 対戦ログ機能の Discord コマンド仕様を定義する。

補足: 本ドキュメントの fighter_id は **Buckler プロフィールの short_id（sid）** を指す。

---

## 1. アカウント連携

### /sf6_link

- 概要: 自分の Buckler short_id（sid）を登録する
- 入力: fighter_id（sid）, display_name (optional)
- 出力: 登録完了メッセージ

### /sf6_unlink

- 概要: 連携解除（収集停止）
- 出力: 解除完了メッセージ

---

## 2. 友達（対戦相手）管理

### /sf6_friend add

- 概要: 対戦相手を登録する
- 入力: fighter_id（sid）, alias (optional)
- 出力: 登録完了メッセージ

### /sf6_friend remove

- 概要: 対戦相手を削除する
- 入力: fighter_id（sid） または alias
- 出力: 削除完了メッセージ

### /sf6_friend list

- 概要: 登録済み友達を一覧する
- 出力: fighter_id / alias の一覧

---

## 3. セッション監視

### /watch_start

- 概要: 指定した友達とのセッション監視を開始する
- 入力: opponent（fighter_id / sid または alias）
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
