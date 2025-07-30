package main

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

type Importer struct {
	archivePath string
	dbPath      string
	verbose     bool
	db          *sql.DB
}

type TweetWrapper struct {
	Tweet struct {
		ID            string `json:"id_str"`
		Text          string `json:"full_text"`
		CreatedAt     string `json:"created_at"`
		RetweetCount  string `json:"retweet_count"`
		FavoriteCount string `json:"favorite_count"`
		Retweeted     bool   `json:"retweeted"`
		Favorited     bool   `json:"favorited"`
		Source        string `json:"source"`
		Lang          string `json:"lang"`
	} `json:"tweet"`
}

type FollowerWrapper struct {
	Follower struct {
		AccountID string `json:"accountId"`
		UserLink  string `json:"userLink"`
	} `json:"follower"`
}

type FollowingWrapper struct {
	Following struct {
		AccountID string `json:"accountId"`
		UserLink  string `json:"userLink"`
	} `json:"following"`
}

type TwitterArchive struct {
	Tweets    []TweetWrapper     `json:"tweets"`
	Followers []FollowerWrapper  `json:"followers"`
	Following []FollowingWrapper `json:"following"`
}

func NewImporter(archivePath, dbPath string, verbose bool) *Importer {
	return &Importer{
		archivePath: archivePath,
		dbPath:      dbPath,
		verbose:     verbose,
	}
}

func (i *Importer) log(format string, args ...interface{}) {
	if i.verbose {
		log.Printf(format, args...)
	}
}

func (i *Importer) Import() error {
	var err error
	i.db, err = sql.Open("duckdb", i.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer i.db.Close()

	if err := i.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	archive, err := i.extractArchive()
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	if err := i.importTweets(archive.Tweets); err != nil {
		return fmt.Errorf("failed to import tweets: %w", err)
	}

	if err := i.importFollowers(archive.Followers); err != nil {
		return fmt.Errorf("failed to import followers: %w", err)
	}

	if err := i.importFollowing(archive.Following); err != nil {
		return fmt.Errorf("failed to import following: %w", err)
	}

	return nil
}

func (i *Importer) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tweets (
			id VARCHAR PRIMARY KEY,
			text TEXT,
			created_at TIMESTAMP,
			retweet_count INTEGER,
			favorite_count INTEGER,
			retweeted BOOLEAN,
			favorited BOOLEAN,
			source TEXT,
			lang VARCHAR(10)
		)`,
		`CREATE TABLE IF NOT EXISTS followers (
			follower_id VARCHAR PRIMARY KEY,
			user_link TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS following (
			following_id VARCHAR PRIMARY KEY,
			user_link TEXT
		)`,
	}

	for _, query := range queries {
		if _, err := i.db.Exec(query); err != nil {
			return err
		}
	}

	i.log("Created database tables")
	return nil
}

func (i *Importer) extractArchive() (*TwitterArchive, error) {
	reader, err := zip.OpenReader(i.archivePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	archive := &TwitterArchive{}
	i.log("Scanning ZIP file contents...")

	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "data/") && strings.HasSuffix(file.Name, ".js") {
			i.log("Found data file: %s", file.Name)
		}
		switch {
		case strings.Contains(file.Name, "tweets.js"):
			i.log("Processing tweets.js file...")
			tweets, err := i.parseTweetsFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse tweets: %w", err)
			}
			archive.Tweets = tweets
			i.log("Parsed %d tweets from file", len(tweets))
		case strings.Contains(file.Name, "follower.js"):
			i.log("Processing follower.js file...")
			followers, err := i.parseFollowersFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse followers: %w", err)
			}
			archive.Followers = followers
			i.log("Parsed %d followers from file", len(followers))
		case strings.Contains(file.Name, "following.js"):
			i.log("Processing following.js file...")
			following, err := i.parseFollowingFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse following: %w", err)
			}
			archive.Following = following
			i.log("Parsed %d following from file", len(following))
		}
	}

	i.log("Extracted archive: %d tweets, %d followers, %d following", 
		len(archive.Tweets), len(archive.Followers), len(archive.Following))
	
	return archive, nil
}

func (i *Importer) parseTweetsFile(file *zip.File) ([]TweetWrapper, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	jsonStart := strings.Index(contentStr, "[")
	if jsonStart == -1 {
		return nil, fmt.Errorf("invalid tweets.js format")
	}

	var tweets []TweetWrapper
	err = json.Unmarshal([]byte(contentStr[jsonStart:]), &tweets)
	return tweets, err
}

func (i *Importer) parseFollowersFile(file *zip.File) ([]FollowerWrapper, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	jsonStart := strings.Index(contentStr, "[")
	if jsonStart == -1 {
		return nil, fmt.Errorf("invalid follower.js format")
	}

	var followers []FollowerWrapper
	err = json.Unmarshal([]byte(contentStr[jsonStart:]), &followers)
	return followers, err
}

func (i *Importer) parseFollowingFile(file *zip.File) ([]FollowingWrapper, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	jsonStart := strings.Index(contentStr, "[")
	if jsonStart == -1 {
		return nil, fmt.Errorf("invalid following.js format")
	}

	var following []FollowingWrapper
	err = json.Unmarshal([]byte(contentStr[jsonStart:]), &following)
	return following, err
}

func (i *Importer) importTweets(tweets []TweetWrapper) error {
	stmt, err := i.db.Prepare(`INSERT OR IGNORE INTO tweets 
		(id, text, created_at, retweet_count, favorite_count, retweeted, favorited, source, lang) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	imported := 0
	for _, tweetWrapper := range tweets {
		tweet := tweetWrapper.Tweet
		if tweet.CreatedAt == "" {
			i.log("Skipping tweet %s: empty created_at", tweet.ID)
			continue
		}

		createdAt, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", tweet.CreatedAt)
		if err != nil {
			i.log("Failed to parse date for tweet %s: %v", tweet.ID, err)
			continue
		}

		retweetCount := 0
		if tweet.RetweetCount != "" {
			fmt.Sscanf(tweet.RetweetCount, "%d", &retweetCount)
		}

		favoriteCount := 0
		if tweet.FavoriteCount != "" {
			fmt.Sscanf(tweet.FavoriteCount, "%d", &favoriteCount)
		}

		_, err = stmt.Exec(
			tweet.ID,
			tweet.Text,
			createdAt,
			retweetCount,
			favoriteCount,
			tweet.Retweeted,
			tweet.Favorited,
			tweet.Source,
			tweet.Lang,
		)
		if err != nil {
			i.log("Failed to insert tweet %s: %v", tweet.ID, err)
		} else {
			imported++
		}
	}

	i.log("Imported %d tweets", imported)
	return nil
}

func (i *Importer) importFollowers(followers []FollowerWrapper) error {
	stmt, err := i.db.Prepare(`INSERT OR IGNORE INTO followers 
		(follower_id, user_link) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	imported := 0
	for _, followerWrapper := range followers {
		follower := followerWrapper.Follower
		_, err = stmt.Exec(follower.AccountID, follower.UserLink)
		if err != nil {
			i.log("Failed to insert follower %s: %v", follower.AccountID, err)
		} else {
			imported++
		}
	}

	i.log("Imported %d followers", imported)
	return nil
}

func (i *Importer) importFollowing(following []FollowingWrapper) error {
	stmt, err := i.db.Prepare(`INSERT OR IGNORE INTO following 
		(following_id, user_link) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	imported := 0
	for _, followingWrapper := range following {
		follow := followingWrapper.Following
		_, err = stmt.Exec(follow.AccountID, follow.UserLink)
		if err != nil {
			i.log("Failed to insert following %s: %v", follow.AccountID, err)
		} else {
			imported++
		}
	}

	i.log("Imported %d following", imported)
	return nil
}