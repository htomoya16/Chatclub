# Chatclub

**Chatclub** は、Street Fighter 6（Buckler）のカスタム対戦ログを取得・集計し、  
Discord Bot から戦績・履歴・セッションを見える化するためのバックエンドである。

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


## 🎥 Demo
https://github.com/user-attachments/assets/793658f9-bb5d-4d9c-b743-d5656dd2a4ce





## 📌 招待URL

Discord Bot をサーバへ招待する URL をここに置く。  

```
https://discord.com/oauth2/authorize?client_id=1461387682172375286&permissions=540142592&integration_type=0&scope=bot+applications.commands
```

## 🧭 セットアップ

セットアップ手順は docs に移動。

- `docs/setup.md`
- `docs/deploy-heroku.md`

## 🧩 コマンド一覧（できること）

### 基本

| コマンド | オプション | 説明 |
|---|---|---|
| `/ping` | なし | Bot の生存確認。 |
| `/anon` | `message` 任意, `file1` 任意, `file2` 任意, `file3` 任意 | 匿名メッセージ投稿（本文・画像添付対応）。 |
| `/anon-channel add` | `channel` 必須 | 匿名投稿を許可するチャンネルを登録。 |
| `/anon-channel remove` | `channel` 必須 | 匿名投稿を許可するチャンネルを解除。 |

`/anon` の入力画面。

![anon command](images/anon-command.png)

`/anon` の投稿結果。

![anon result](images/anon-result.png)

### SF6（Buckler）
※ SF6系コマンドは Street Fighter 6 のアカウント連携が必要。  
未連携の場合は使用できない。  
Buckler: https://www.streetfighter.com/6/buckler/ja-jp
また、以下コマンドで使うユーザーコード（sid）とは Buckler's Boot Camp のプロフィールサイトに書いてあるユーザーコード。

![bucker profile](images/buckler-profile.png)


#### SF6 コマンドで使う主なオプション

| オプション | 使うコマンド | 説明 |
|---|---|---|
| `opponent_code` | `/sf6_stats`, `/sf6_history`, `/sf6_session` | 対戦相手側の Buckler ユーザーコード（sid）。たとえば「自分 vs 相手」の戦績を見たい場合は、相手のプロフィールに表示されているユーザーコードを入れる。必須。相手がこのサーバーでSF6連携済みなら Discord のメンションでも指定できる。 |
| `subject_code` | `/sf6_stats`, `/sf6_history`, `/sf6_session` | 集計対象にする自分側の Buckler ユーザーコード（sid）。任意。未指定ならコマンド実行者の連携アカウントを使う。別の連携済みユーザーを集計対象にしたい場合は Discord のメンションでも指定できる。 |
| `from` | `/sf6_stats range` | 開始日。必須。`YYYY-MM-DD` 形式、JST。 |
| `to` | `/sf6_stats range` | 終了日。必須。`YYYY-MM-DD` 形式、JST。 |
| `count` | `/sf6_stats count` | 集計する直近試合数。必須。 |

例: 自分の連携アカウントと相手 `1234567890` の直近20戦を見る場合は、`/sf6_stats count opponent_code:1234567890 count:20` のように入力する。  
例: `@PlayerA` と `@PlayerB` がどちらもSF6連携済みなら、`subject_code:@PlayerA opponent_code:@PlayerB` のようにメンションでも指定できる。

#### SF6 コマンド一覧

| コマンド | オプション | 説明 |
|---|---|---|
| `/sf6_account` | なし | 連携状況の表示・連携/解除ボタンの提示。自身の Street Fighter 6 アカウントを連携・解除できる。 |
| `/sf6_unlink` | なし | 自身の Street Fighter 6 アカウント連携を解除する。戦績データは保持される。 |
| `/sf6_friend` | なし | フレンド一覧と追加/削除。フレンドの Street Fighter 6 アカウントを連携できる。 |
| `/sf6_fetch` | なし | 対戦ログの手動取得（管理者/許可ユーザー）。 |
| `/sf6_stats range` | `opponent_code` 必須, `from` 必須, `to` 必須, `subject_code` 任意 | 期間指定の戦績集計（JST）。 |
| `/sf6_stats count` | `opponent_code` 必須, `count` 必須, `subject_code` 任意 | 直近N戦の勝率などを集計。 |
| `/sf6_stats set` | `opponent_code` 必須, `subject_code` 任意 | 連戦を1セットとして勝率などを集計（30分以内の試合間隔を同一セット扱い）。 |
| `/sf6_history` | `opponent_code` 必須, `subject_code` 任意 | 対戦履歴の一覧表示（ページング）。 |
| `/sf6_session start` | `opponent_code` 必須, `subject_code` 任意 | セッション開始。 |
| `/sf6_session end` | `opponent_code` 必須, `subject_code` 任意 | セッション終了と集計。end時にそのセッション内の対戦だけをまとめて集計し、戦績を表示。 |

`/sf6_account` の表示。

![sf6 account](images/sf6-account.png)

`/sf6_friend` の表示。

![sf6 friend](images/sf6-friend.png)

`/sf6_stats range` の表示。

![sf6 stats range](images/sf6-stats-range.png)

`/sf6_stats count` の表示。

![sf6 stats count](images/sf6-stats-count.png)

`/sf6_stats set` の表示。

![sf6 stats set](images/sf6-stats-set.png)

`/sf6_history` の表示。

![sf6 history](images/sf6-history.png)

