# SF6 Buckler 対戦ログ 概要

本機能は、**Street Fighter 6 の Buckler’s Boot Camp** にある対戦履歴から、
「自分 vs 登録済みの友達」の**カスタムマッチ戦績**を自動収集し、
Discord で勝率・傾向・セッション単位の統計を可視化するための仕組みである。

---

## 1. 目的

- 特定の友達とのカスタムマッチ戦績を自動収集する
- 勝率や苦手な組み合わせなどの**傾向**を継続的に把握する
- /watch_start での**セッション監視**によりリアルタイム寄りの集計を提供する

---

## 2. 対象データと前提

- **Buckler の Battle Log（JSON 相当）**を一定間隔で取得する
- 対象は **Custom Room の試合のみ**
- 「自分（登録済みアカウント） vs 登録友達」の試合だけを保存対象とする
- 取得漏れ対策として**直近 N 件の再取得 + 重複排除**を必須とする

補足（取得仕様）:

- Buckler は **Next.js の data API** から JSON を返す
- Custom の取得先は以下（`buildId` は HTML の `__NEXT_DATA__.buildId` から抽出）
- `/6/buckler/_next/data/{buildId}/ja-jp/profile/{sid}/battlelog/custom.json?sid={sid}&page={page}`（sid＝ユーザーコード）
- 実データは `pageProps.replay_list`
- ページングは `page` クエリ、1ページ 10 件
- 認証はログイン済み Cookie に依存する（`buckler_id` / `buckler_r_id` など）
- buildId は更新で変わるため **キャッシュ + 失効時再取得**を行う
- 自動ログインで Cookie を更新する（失効時は再ログイン）
- プレイヤーIDは `player.*.short_id`（`sid`（ユーザーコード））、表示名は `player.*.fighter_id`

---

## 3. 機能範囲

### 3.1 友達（対戦相手）登録

- CFN 識別子（fighter_id 等）で友達を登録する
- 覚えやすい別名（alias）を付けられる

### 3.2 Buckler ポーリング

- 一定間隔で Battle Log を取得する
- Custom Room かつ「自分 vs 登録友達」に該当する試合のみ保存する
- 直近 N 件の再取得により欠落を補完する

### 3.3 セッション監視

- `/watch_start opponent=<友達>` でセッション監視を開始する
- 監視中は短い間隔でポーリングし、**開始時刻以降**の増分試合を取得する
- `/watch_end` または一定時間更新なしで自動終了する
- 終了時に直近 N 件を再取得して取りこぼしを補完する

### 3.4 統計表示

- 通算：勝率 / 総試合数 / 直近 N 戦勝率 / 連勝・連敗
- キャラ別：自キャラ × 相手キャラの勝率
- セッション単位：セッション中の勝敗推移・連勝/連敗
- 期間推移（日/週/月）は必要に応じて追加する

---

## 4. 非スコープ

- ランクマ / カジュアルなど **Custom Room 以外の試合**の集計
- 試合内容の詳細分析（フレーム・ダメージ等）
- Buckler 側の規約に反する取得方法

---

## 5. 関連ドキュメント

- `docs/sf6-buckler/command.md`
- `docs/sf6-buckler/flow.md`
- `docs/sf6-buckler/data-model.md`
- `docs/sf6-buckler/stats.md`
