# AGENTS.md

本リポジトリは、Discord サーバメンバーの  
**通話活動（VC）・ゲーム活動（Steam 等）を記録・集計し、Web で可視化する Discord Bot + Web サービス**である。

本ファイル（AGENTS.md）は **プロジェクト全体の方針・ルール・責務分離のみ**を定義する。  
**各機能の詳細仕様は docs/ 以下に記述すること。**

---

## 1. このリポジトリにおける役割分担

- **AGENTS.md**
  - プロジェクトの目的
  - ディレクトリ責務
  - 実装ルール・設計原則
  - docs/ への導線
- **docs/**
  - 各機能の仕様書
  - データ設計・フロー・UI 仕様
  - 拡張案・未実装機能の設計メモ

AGENTS.md に機能仕様を書き足してはならない。

---

## 2. ディレクトリ責務（固定）

```

/
cmd/          エントリポイント（bot / api / worker 等）
config/       設定ロード・環境変数・バリデーション
database/     DB 接続・Repository・Tx 管理
internal/     ドメインロジック（Bot / Stats / Remind 等）
migrations/   DB マイグレーション（Atlas）
schema/       現行 DB スキーマ定義
scripts/      開発・運用補助スクリプト
docs/         機能仕様・設計ドキュメント（最重要）

```

---

## 3. docs/ の構成ルール

docs/ 配下は **機能単位で分割**する。

例：

```

docs/
vc-log/
overview.md
data-model.md
flow.md

game-activity/
overview.md
providers.md
ranking.md

remind/
overview.md
command.md
scheduler.md

anonymous-chat/
overview.md

web/
overview.md
stats-pages.md

roadmap.md

```

- `overview.md` は必須
- 実装前・検討中の機能も docs に書いてよい
- docs に書かれていない機能は「未定義」とみなす

---

## 4. 実装原則（重要）

### 4.1 責務分離
- Discord API 直接操作は `internal/discordbot` に閉じる
- ビジネスロジックは Discord 非依存で実装する
- Web / Bot / Worker は **同じドメインロジックを共有**する

### 4.2 データ設計
- 永続化するデータ構造は **必ず docs → migrations の順で決める**
- DB スキーマをコード先行で決めない

### 4.3 時刻と集計
- DB 保存は UTC
- 集計ロジックは純関数に近づけ、テスト可能にする

---

## 5. 開発フロー

1. 機能を追加・変更したい場合
   - docs/ に仕様を書く
2. 仕様レビュー（自分 or 将来の自分）
3. DB 変更があれば migrations を更新
4. 実装（internal → cmd）
5. docs と実装の不整合が出た場合
   - **docs を正とする**

---

## 6. 非スコープ管理

以下のような機能は **docs に設計だけ置くことは許可**するが、
AGENTS.md や core 実装に混ぜない。

- AI チャット（API 料金が発生するもの）
- 一言 AI コメント
- ミニゲーム（大富豪など）
- 高度な外部 API 連携（Valorant など）

---

## 7. 最低限の品質基準

- 集計・ランキング系ロジックはテスト可能であること
- 失敗しても Bot が落ちない設計にする
- Discord ID / Guild ID を唯一の識別子として扱う

---

## 8. このファイルの扱い

- AGENTS.md は「憲法」であり、頻繁に書き換えない
- 迷ったら docs/ に逃がす
- 機能の増殖は docs/ で管理する

```


