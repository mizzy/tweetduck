# TweetDuck

Twitter Archive to DuckDB Importer

TweetDuckは、TwitterアーカイブのZIPファイルからデータを抽出し、DuckDBデータベースにインポートするGoプログラムです。

## 機能

- TwitterアーカイブZIPファイルの自動解析
- ツイート、リプライ、お気に入り、フォロワーなどのデータ抽出
- DuckDBへの高速データインポート
- 重複データの自動処理

## 必要要件

- Go 1.21以上
- DuckDB

## インストール

```bash
go mod init tweetduck
go mod tidy
go build -o tweetduck
```

## 使用方法

```bash
./tweetduck -archive twitter-archive.zip -db output.duckdb
```

### オプション

- `-archive`: TwitterアーカイブのZIPファイルパス（必須）
- `-db`: 出力先DuckDBファイルパス（デフォルト: tweets.duckdb）
- `-verbose`: 詳細ログの出力

## データベーススキーマ

### tweets テーブル
- id: ツイートID
- text: ツイート本文
- created_at: 作成日時
- retweet_count: リツイート数
- favorite_count: お気に入り数
- user_id: ユーザーID
- user_name: ユーザー名
- user_screen_name: ユーザースクリーンネーム

### followers テーブル
- follower_id: フォロワーID
- follower_screen_name: フォロワースクリーンネーム
- followed_at: フォロー日時

### following テーブル
- following_id: フォロー中のユーザーID
- following_screen_name: フォロー中のユーザースクリーンネーム
- followed_at: フォロー開始日時

## ライセンス

MIT License