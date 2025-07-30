# TweetDuck

Twitter Archive to DuckDB Importer

TweetDuckは、TwitterアーカイブのZIPファイルからデータを抽出し、DuckDBデータベースにインポートするGoプログラムです。2025年形式の新しいTwitterアーカイブに対応しています。

## 機能

- **2025年形式対応**: 最新のTwitterアーカイブZIPファイルの自動解析
- **包括的データ抽出**: ツイート、フォロワー、フォロイングデータの抽出
- **高速インポート**: DuckDBへの効率的なデータインポート
- **重複データ処理**: INSERT OR IGNOREによる重複データの自動処理
- **詳細ログ**: verboseモードでの詳細な処理状況表示

## 必要要件

- Go 1.21以上
- DuckDBライブラリ（自動取得）

## インストール

```bash
git clone <repository-url>
cd tweetduck
go mod tidy
go build -o tweetduck
```

## 使用方法

### 基本的な使用方法
```bash
./tweetduck --archive="path/to/twitter-archive.zip" --db="output.duckdb"
```

### 詳細ログ付きで実行
```bash
./tweetduck --archive="path/to/twitter-archive.zip" --db="output.duckdb" --verbose
```

### オプション

- `--archive`, `-a`: TwitterアーカイブのZIPファイルパス（必須）
- `--db`, `-d`: 出力先DuckDBファイルパス（デフォルト: tweets.duckdb）
- `--verbose`, `-v`: 詳細ログの出力

## データベーススキーマ

### tweets テーブル
- `id` (VARCHAR): ツイートID（主キー）
- `text` (TEXT): ツイート本文
- `created_at` (TIMESTAMP): 作成日時
- `retweet_count` (INTEGER): リツイート数
- `favorite_count` (INTEGER): お気に入り数
- `retweeted` (BOOLEAN): リツイート済みフラグ
- `favorited` (BOOLEAN): お気に入り済みフラグ
- `source` (TEXT): 投稿元アプリケーション
- `lang` (VARCHAR(10)): 言語コード

### followers テーブル
- `follower_id` (VARCHAR): フォロワーのユーザーID（主キー）
- `user_link` (TEXT): フォロワーのTwitterリンク

### following テーブル
- `following_id` (VARCHAR): フォロー中のユーザーID（主キー）
- `user_link` (TEXT): フォロー中ユーザーのTwitterリンク

## 使用例

### データ検索例

```sql
-- ドクターペッパー関連のツイートを検索
SELECT text, created_at FROM tweets 
WHERE text LIKE '%ドクペ%' OR text LIKE '%ドクターペッパー%' OR text LIKE '%Dr Pepper%'
ORDER BY created_at DESC 
LIMIT 10;

-- 特定の期間のツイート数をカウント
SELECT DATE(created_at) as date, COUNT(*) as tweet_count 
FROM tweets 
WHERE created_at >= '2025-01-01' 
GROUP BY DATE(created_at) 
ORDER BY date DESC;

-- フォロワー数とフォロー数を確認
SELECT 
  (SELECT COUNT(*) FROM followers) as follower_count,
  (SELECT COUNT(*) FROM following) as following_count;
```

## 対応アーカイブ形式

- 2025年形式のTwitterアーカイブ（新形式）
- `data/tweets.js` および `data/deleted-tweets.js` の処理
- `data/follower.js` および `data/following.js` の処理
- JavaScript形式データファイル（`window.YTD.*` 形式）

## ライセンス

MIT License