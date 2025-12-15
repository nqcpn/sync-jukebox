package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// 定义数据结构
type Song struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	DurationMs int    `json:"duration_ms"`
	Source     string `json:"source"` // "local", "netease"
	FilePath   string `json:"-"`      // 不暴露给前端
}

type PlaylistItem struct {
	ID     int    `json:"id"`
	SongID string `json:"song_id"`
	Order  int    `json:"order"`
	Song   *Song  `json:"song,omitempty"` // 关联的歌曲信息
}

// DB 是数据库操作的封装
type DB struct {
	*sql.DB
}

// New 初始化并返回一个数据库连接
func New(dataSourceName string) (*DB, error) {
	database, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{database}
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema 创建所有必要的表
func (db *DB) initSchema() error {
	// 使用 TEXT 作为主键，因为我们将使用UUID
	createSongsTable := `
	CREATE TABLE IF NOT EXISTS songs (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		artist TEXT,
		album TEXT,
		duration_ms INTEGER,
		source TEXT,
		file_path TEXT NOT NULL UNIQUE
	);`

	createPlaylistItemsTable := `
	CREATE TABLE IF NOT EXISTS playlist_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		song_id TEXT NOT NULL,
		item_order INTEGER,
		FOREIGN KEY(song_id) REFERENCES songs(id) ON DELETE CASCADE
	);`

	createTokensTable := `
	CREATE TABLE IF NOT EXISTS tokens (
		token TEXT PRIMARY KEY,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 用于持久化播放器状态
	createSystemStateTable := `
	CREATE TABLE IF NOT EXISTS system_state (
		key TEXT PRIMARY KEY,
		value TEXT
	);`

	for _, stmt := range []string{createSongsTable, createPlaylistItemsTable, createTokensTable, createSystemStateTable} {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// --- Token 操作 ---
func (db *DB) IsTokenValid(token string) (bool, error) {
	var isActive bool
	err := db.QueryRow("SELECT is_active FROM tokens WHERE token = ?", token).Scan(&isActive)
	if err == sql.ErrNoRows {
		return false, nil // Token不存在
	}
	return isActive, err
}

func (db *DB) AddToken(token string) error {
	_, err := db.Exec("INSERT INTO tokens (token) VALUES (?)", token)
	return err
}

func (db *DB) SetTokenState(token string, isActive bool) error {
	_, err := db.Exec("UPDATE tokens SET is_active = ? WHERE token = ?", isActive, token)
	return err
}

// --- System State 操作 ---
func (db *DB) GetSystemState(key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM system_state WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil // Key不存在是正常情况
	}
	return value, err
}

func (db *DB) SetSystemState(key, value string) error {
	// "INSERT OR REPLACE" 是一种便捷的 upsert 写法
	_, err := db.Exec("INSERT OR REPLACE INTO system_state (key, value) VALUES (?, ?)", key, value)
	return err
}

// --- Song 操作 ---
func (db *DB) AddSong(song *Song) error {
	_, err := db.Exec("INSERT INTO songs (id, title, artist, album, duration_ms, source, file_path) VALUES (?, ?, ?, ?, ?, ?, ?)",
		song.ID, song.Title, song.Artist, song.Album, song.DurationMs, song.Source, song.FilePath)
	return err
}

func (db *DB) GetSong(id string) (*Song, error) {
	s := &Song{}
	err := db.QueryRow("SELECT id, title, artist, album, duration_ms, source, file_path FROM songs WHERE id = ?", id).Scan(
		&s.ID, &s.Title, &s.Artist, &s.Album, &s.DurationMs, &s.Source, &s.FilePath)
	return s, err
}

func (db *DB) GetAllSongs() ([]Song, error) {
	rows, err := db.Query("SELECT id, title, artist, album, duration_ms, source, file_path FROM songs ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []Song
	for rows.Next() {
		var s Song
		if err := rows.Scan(&s.ID, &s.Title, &s.Artist, &s.Album, &s.DurationMs, &s.Source, &s.FilePath); err != nil {
			return nil, err
		}
		songs = append(songs, s)
	}
	return songs, nil
}

// DeleteSong 从数据库中删除一首歌曲
func (db *DB) DeleteSong(id string) error {
	_, err := db.Exec("DELETE FROM songs WHERE id = ?", id)
	return err
}

// --- Playlist 操作 ---
func (db *DB) GetPlaylistItems() ([]PlaylistItem, error) {
	rows, err := db.Query("SELECT id, song_id, item_order FROM playlist_items ORDER BY item_order")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PlaylistItem
	for rows.Next() {
		var item PlaylistItem
		if err := rows.Scan(&item.ID, &item.SongID, &item.Order); err != nil {
			return nil, err
		}
		// 填充关联的歌曲信息
		song, err := db.GetSong(item.SongID)
		if err != nil {
			log.Printf("Warning: song %s in playlist not found in library, skipping: %v", item.SongID, err)
			continue
		}
		item.Song = song
		items = append(items, item)
	}
	return items, nil
}

// UpdatePlaylist 完全重写播放列表，确保顺序正确
func (db *DB) UpdatePlaylist(songIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 事务处理，确保原子性
	_, err = tx.Exec("DELETE FROM playlist_items")
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO playlist_items (song_id, item_order) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i, songID := range songIDs {
		_, err := stmt.Exec(songID, i)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
