// tracking/sqlite_tracker.go
package tracking

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// IsCgoEnabled indicates whether CGO is enabled for SQLite support
var IsCgoEnabled = true

// Error constants
var (
	ErrCgoDisabled      = errors.New("CGO disabled: sqlite tracking not available")
	ErrTrackerNotInited = errors.New("tracker not initialized")
)

/*
────────────────────────────────────────────────────────────────────────────*
│  Constantes de Configuração                                                │
*────────────────────────────────────────────────────────────────────────────
*/
const (
	defaultCacheSize  = -20000    // 20MB
	mmapSize          = 268435456 // 256MB
	busyTimeout       = 5000      // 5 segundos
	walAutoCheckpoint = 1000      // páginas
	maxOpenConns      = 5         // conexões simultâneas
	maxIdleConns      = 2         // conexões inativas
	avgAnimePerUser   = 100       // pré-alocação de slices
)

/*
────────────────────────────────────────────────────────────────────────────*
│  Tipos e Estruturas                                                        │
*────────────────────────────────────────────────────────────────────────────
*/
type Anime struct {
	AnilistID     int       `json:"anilist_id"`
	AllanimeID    string    `json:"allanime_id"`
	EpisodeNumber int       `json:"episode_number"`
	PlaybackTime  int       `json:"playback_time"`
	Duration      int       `json:"duration"`
	Title         string    `json:"title"`
	LastUpdated   time.Time `json:"last_updated"`
}

type LocalTracker struct {
	db       *sql.DB
	upsertPS *sql.Stmt
	getPS    *sql.Stmt
	allPS    *sql.Stmt
	deletePS *sql.Stmt
}

/*
────────────────────────────────────────────────────────────────────────────*
│  Construtor e Inicialização                                                │
*────────────────────────────────────────────────────────────────────────────
*/
var NewLocalTracker func(dbPath string) *LocalTracker

func newLocalTrackerImpl(dbPath string) *LocalTracker {
	// Check if CGO is disabled (SQLite not available)
	if !IsCgoEnabled {
		fmt.Println("Warning: CGO is disabled, anime progress tracking will be unavailable")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		return nil
	}
	// No Windows, os caminhos precisam ser tratados de forma especial para o SQLite
	var dsn string
	if runtime.GOOS == "windows" {
		// Usar URI format com escape para Windows
		escapedPath := strings.ReplaceAll(dbPath, "\\", "/")
		dsn = fmt.Sprintf(
			"file:%s?_journal_mode=WAL&_synchronous=NORMAL&_wal_autocheckpoint=%d&"+
				"_busy_timeout=%d&_cache_size=%d&_mmap_size=%d&_mode=rwc",
			escapedPath,
			walAutoCheckpoint,
			busyTimeout,
			defaultCacheSize,
			mmapSize,
		)
	} else {
		dsn = fmt.Sprintf(
			"file:%s?_journal_mode=WAL&_synchronous=NORMAL&_wal_autocheckpoint=%d&"+
				"_busy_timeout=%d&_cache_size=%d&_mmap_size=%d",
			dbPath,
			walAutoCheckpoint,
			busyTimeout,
			defaultCacheSize,
			mmapSize,
		)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return nil
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	if err := initializeDatabase(db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Printf("Error closing database: %v\n", closeErr)
		}
		fmt.Printf("Error initializing database: %v\n", err)
		return nil
	}

	statements, err := prepareStatements(db)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Printf("Error closing database: %v\n", closeErr)
		}
		fmt.Printf("Error preparing statements: %v\n", err)
		return nil
	}

	return &LocalTracker{
		db:       db,
		upsertPS: statements.upsert,
		getPS:    statements.get,
		allPS:    statements.all,
		deletePS: statements.delete,
	}
}

/*
────────────────────────────────────────────────────────────────────────────*
│  Inicialização do Banco de Dados                                           │
*────────────────────────────────────────────────────────────────────────────
*/
func initializeDatabase(db *sql.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS anime_progress (
		anilist_id     INTEGER NOT NULL,
		allanime_id    TEXT    NOT NULL,
		episode_number INTEGER NOT NULL,
		playback_time  INTEGER NOT NULL CHECK(playback_time >= 0),
		duration       INTEGER NOT NULL CHECK(duration > 0),
		title          TEXT,
		last_updated   INTEGER NOT NULL,
		PRIMARY KEY (anilist_id, allanime_id)
	);`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("schema creation failed: %w", err)
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_anime_cover 
		ON anime_progress(
			anilist_id,
			allanime_id,
			episode_number,
			playback_time,
			duration,
			title,
			last_updated
		)`,
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("index creation '%s' failed: %w", idx, err)
		}
	}

	if _, err := db.Exec(`PRAGMA optimize`); err != nil {
		return fmt.Errorf("initial optimization failed: %w", err)
	}

	return nil
}

/*
────────────────────────────────────────────────────────────────────────────*
│  Preparação de Statements                                                  │
*────────────────────────────────────────────────────────────────────────────
*/
type preparedStatements struct {
	upsert *sql.Stmt
	get    *sql.Stmt
	all    *sql.Stmt
	delete *sql.Stmt
}

func prepareStatements(db *sql.DB) (*preparedStatements, error) {
	upsert, err := db.Prepare(`INSERT INTO anime_progress (
		anilist_id, 
		allanime_id, 
		episode_number, 
		playback_time, 
		duration, 
		title, 
		last_updated
	) VALUES (?,?,?,?,?,?,?) 
	ON CONFLICT(anilist_id, allanime_id) DO UPDATE SET
		episode_number = excluded.episode_number,
		playback_time = excluded.playback_time,
		duration = excluded.duration,
		title = excluded.title,
		last_updated = excluded.last_updated`)

	if err != nil {
		return nil, fmt.Errorf("upsert preparation failed: %w", err)
	}

	get, err := db.Prepare(`SELECT 
		episode_number, 
		playback_time, 
		duration, 
		title, 
		last_updated 
	FROM anime_progress 
	WHERE anilist_id = ? AND allanime_id = ?`)

	if err != nil {
		return nil, fmt.Errorf("get preparation failed: %w", err)
	}

	all, err := db.Prepare(`SELECT 
		anilist_id, 
		allanime_id, 
		episode_number, 
		playback_time, 
		duration, 
		title, 
		last_updated 
	FROM anime_progress`)

	if err != nil {
		return nil, fmt.Errorf("all preparation failed: %w", err)
	}

	delete, err := db.Prepare(`DELETE FROM anime_progress 
		WHERE anilist_id = ? AND allanime_id = ?`)

	if err != nil {
		return nil, fmt.Errorf("delete preparation failed: %w", err)
	}

	return &preparedStatements{
		upsert: upsert,
		get:    get,
		all:    all,
		delete: delete,
	}, nil
}

/*
────────────────────────────────────────────────────────────────────────────*
│  Operações Principais                                                      │
*────────────────────────────────────────────────────────────────────────────
*/
func (t *LocalTracker) UpdateProgress(a Anime) error {
	// Safety check for when tracker is not initialized
	if t == nil || t.db == nil || t.upsertPS == nil {
		return ErrTrackerNotInited
	}

	// Validate duration to prevent constraint errors
	if a.Duration <= 0 {
		return fmt.Errorf("invalid duration value (%d): must be greater than 0", a.Duration)
	}

	// Validate playback time (shouldn't be negative)
	if a.PlaybackTime < 0 {
		a.PlaybackTime = 0
	}

	_, err := t.upsertPS.Exec(
		a.AnilistID,
		a.AllanimeID,
		a.EpisodeNumber,
		a.PlaybackTime,
		a.Duration,
		a.Title,
		a.LastUpdated.Unix(),
	)
	return err
}

func (t *LocalTracker) GetAnime(anilistID int, allanimeID string) (*Anime, error) {
	// Safety check for when tracker is not initialized
	if t == nil || t.db == nil || t.getPS == nil {
		return nil, ErrTrackerNotInited
	}

	var a Anime
	var ts int64

	err := t.getPS.QueryRow(anilistID, allanimeID).Scan(
		&a.EpisodeNumber,
		&a.PlaybackTime,
		&a.Duration,
		&a.Title,
		&ts,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	a.AnilistID = anilistID
	a.AllanimeID = allanimeID
	a.LastUpdated = time.Unix(ts, 0)
	return &a, nil
}

func (t *LocalTracker) GetAllAnime() ([]Anime, error) {
	// Safety check for when tracker is not initialized
	if t == nil || t.db == nil || t.allPS == nil {
		return nil, ErrTrackerNotInited
	}

	rows, err := t.allPS.Query()
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	list := make([]Anime, 0, avgAnimePerUser)
	for rows.Next() {
		var a Anime
		var ts int64
		if err := rows.Scan(
			&a.AnilistID,
			&a.AllanimeID,
			&a.EpisodeNumber,
			&a.PlaybackTime,
			&a.Duration,
			&a.Title,
			&ts,
		); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		a.LastUpdated = time.Unix(ts, 0)
		list = append(list, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return list, nil
}

func (t *LocalTracker) DeleteAnime(anilistID int, allanimeID string) error {
	_, err := t.deletePS.Exec(anilistID, allanimeID)
	return err
}

/*
────────────────────────────────────────────────────────────────────────────*
│  Finalização                                                               │
*────────────────────────────────────────────────────────────────────────────
*/
func (t *LocalTracker) Close() error {
	var finalErr error

	closeStmt := func(stmt *sql.Stmt, name string) {
		if stmt != nil {
			if err := stmt.Close(); err != nil {
				finalErr = fmt.Errorf("%s statement close error: %w", name, err)
			}
		}
	}

	closeStmt(t.upsertPS, "upsert")
	closeStmt(t.getPS, "get")
	closeStmt(t.allPS, "all")
	closeStmt(t.deletePS, "delete")

	if err := t.db.Close(); err != nil {
		finalErr = fmt.Errorf("database close error: %w", err)
	}

	return finalErr
}

func init() {
	// This will be replaced at build time with false if CGO is disabled
	// When using CGO_ENABLED=0
	IsCgoEnabled = isCgoEnabled()
	NewLocalTracker = newLocalTrackerImpl // Initialize the public variable
}

// The implementation of isCgoEnabled is defined in local_cgo.go and local_nocgo.go
// based on build tags. We don't define it here to avoid duplicate declarations.
