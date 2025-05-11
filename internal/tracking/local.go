// // tracking/sqlite_tracker.go
// //
// // Tracker que usa **SQLite em modo WAL**.
// // • Cada update executa um INSERT … ON CONFLICT DO UPDATE e retorna
// //   somente depois de a transação estar no disco (durável).
// // • Sem goroutines nem filas; latência típica < 1 ms em SSD.
// // • DSN: journal_mode=WAL, synchronous=NORMAL, busy_timeout=3000 ms.

// package tracking

// import (
// 	"database/sql"
// 	"fmt"
// 	"time"

// 	_ "github.com/mattn/go-sqlite3"
// )

// /*────────────────────────────────────────────────────────────────────────────*
//  |  Estruturas                                                               |
//  *────────────────────────────────────────────────────────────────────────────*/

// type Anime struct {
// 	AnilistID     int
// 	AllanimeID    string
// 	EpisodeNumber int
// 	PlaybackTime  int
// 	Duration      int
// 	Title         string
// 	LastUpdated   time.Time
// }

// type LocalTracker struct {
// 	db *sql.DB
// }

// /*────────────────────────────────────────────────────────────────────────────*
//  |  Inicialização                                                            |
//  *────────────────────────────────────────────────────────────────────────────*/

// // NewLocalTracker abre ou cria o banco e garante o schema.
// func NewLocalTracker(dbPath string) *LocalTracker {
// 	dsn := fmt.Sprintf(
// 		"file:%s?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=3000&_foreign_keys=on",
// 		dbPath,
// 	)
// 	db, err := sql.Open("sqlite3", dsn)
// 	if err != nil {
// 		panic(err) // deixe propagar ou trate conforme seu projeto
// 	}
// 	db.SetMaxOpenConns(1)

// 	const schema = `CREATE TABLE IF NOT EXISTS anime_progress (
// 		anilist_id     INTEGER NOT NULL,
// 		allanime_id    TEXT    NOT NULL,
// 		episode_number INTEGER NOT NULL,
// 		playback_time  INTEGER NOT NULL,
// 		duration       INTEGER NOT NULL,
// 		title          TEXT,
// 		last_updated   TEXT    NOT NULL,
// 		PRIMARY KEY (anilist_id, allanime_id)
// 	);`
// 	if _, err = db.Exec(schema); err != nil {
// 		db.Close()
// 		panic(err)
// 	}
// 	return &LocalTracker{db: db}
// }

// /*────────────────────────────────────────────────────────────────────────────*
//  |  Operações públicas                                                       |
//  *────────────────────────────────────────────────────────────────────────────*/

// // UpdateProgress grava (upsert) imediatamente.
// func (t *LocalTracker) UpdateProgress(a Anime) error {
// 	a.LastUpdated = time.Now().UTC()

// 	const upsert = `
// 	INSERT INTO anime_progress
// 	    (anilist_id, allanime_id, episode_number,
// 	     playback_time, duration, title, last_updated)
// 	VALUES (?,?,?,?,?,?,?)
// 	ON CONFLICT(anilist_id, allanime_id) DO UPDATE SET
// 		episode_number = excluded.episode_number,
// 		playback_time  = excluded.playback_time,
// 		duration       = excluded.duration,
// 		title          = excluded.title,
// 		last_updated   = excluded.last_updated;
// 	`
// 	_, err := t.db.Exec(upsert,
// 		a.AnilistID, a.AllanimeID,
// 		a.EpisodeNumber, a.PlaybackTime,
// 		a.Duration, a.Title, a.LastUpdated.Format(time.RFC3339),
// 	)
// 	return err
// }

// // DeleteAnime remove um registro.
// func (t *LocalTracker) DeleteAnime(anilistID int, allanimeID string) error {
// 	_, err := t.db.Exec(
// 		`DELETE FROM anime_progress
// 		  WHERE anilist_id=? AND allanime_id=?`,
// 		anilistID, allanimeID,
// 	)
// 	return err
// }

// // GetAnime retorna um registro específico.
// func (t *LocalTracker) GetAnime(anilistID int, allanimeID string) (*Anime, error) {
// 	row := t.db.QueryRow(
// 		`SELECT anilist_id, allanime_id, episode_number,
// 		        playback_time, duration, title, last_updated
// 		   FROM anime_progress
// 		  WHERE anilist_id=? AND allanime_id=?`,
// 		anilistID, allanimeID,
// 	)
// 	var a Anime
// 	var ts string
// 	err := row.Scan(&a.AnilistID, &a.AllanimeID,
// 		&a.EpisodeNumber, &a.PlaybackTime,
// 		&a.Duration, &a.Title, &ts)
// 	if err == sql.ErrNoRows {
// 		return nil, nil
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	a.LastUpdated, _ = time.Parse(time.RFC3339, ts)
// 	return &a, nil
// }

// // GetAllAnime devolve slice com todos registros.
// func (t *LocalTracker) GetAllAnime() ([]Anime, error) {
// 	rows, err := t.db.Query(
// 		`SELECT anilist_id, allanime_id, episode_number,
// 		        playback_time, duration, title, last_updated
// 		   FROM anime_progress`,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var list []Anime
// 	for rows.Next() {
// 		var a Anime
// 		var ts string
// 		if err = rows.Scan(&a.AnilistID, &a.AllanimeID,
// 			&a.EpisodeNumber, &a.PlaybackTime,
// 			&a.Duration, &a.Title, &ts); err != nil {
// 			return nil, err
// 		}
// 		a.LastUpdated, _ = time.Parse(time.RFC3339, ts)
// 		list = append(list, a)
// 	}
// 	return list, rows.Err()
// }

// // Close fecha a conexão.
// func (t *LocalTracker) Close() error { return t.db.Close() }


// tracking/sqlite_tracker.go
//
// Rastreador de progresso usando **SQLite + WAL** otimizado:
// ─ Prepared‑statement único para UPSERT  → menos “prepare/step”.
// ─ PRAGMAs de performance (_wal_autocheckpoint, _cache_size, busy_timeout).
// ─ Cada operação grava em < 1 ms (SSD) e permanece durável.
//
// Requer: `go get github.com/mattn/go-sqlite3` (CGO).

// tracking/sqlite_tracker.go
//
// Rastreador de progresso usando **SQLite + WAL** otimizado:
// ─ Prepared‑statement único para UPSERT  → menos “prepare/step”.
// ─ PRAGMAs de performance (_wal_autocheckpoint, _cache_size, busy_timeout).
// ─ Cada operação grava em < 1 ms (SSD) e permanece durável.
//
// Requer: `go get github.com/mattn/go-sqlite3` (CGO).

package tracking

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

/*────────────────────────────────────────────────────────────────────────────*
 |  Tipos                                                                    |
 *────────────────────────────────────────────────────────────────────────────*/

// Anime é o payload que o app usa.
type Anime struct {
	AnilistID     int
	AllanimeID    string
	EpisodeNumber int
	PlaybackTime  int
	Duration      int
	Title         string
	LastUpdated   time.Time
}

// LocalTracker mantém a conexão e o statement preparado.
type LocalTracker struct {
	db        *sql.DB
	upsertPS  *sql.Stmt
}

/*────────────────────────────────────────────────────────────────────────────*
 |  Construtor — abre ou cria DB e prepara o statement                       |
 *────────────────────────────────────────────────────────────────────────────*/

func NewLocalTracker(dbPath string) *LocalTracker {
	dsn := fmt.Sprintf(
		`file:%s?
		  _journal_mode=WAL&
		  _synchronous=NORMAL&
		  _wal_autocheckpoint=1000&
		  _busy_timeout=3000&
		  _cache_size=-20000`,
		dbPath,
	)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1) // 1 writer é o bastante

	const schema = `CREATE TABLE IF NOT EXISTS anime_progress (
		anilist_id     INTEGER NOT NULL,
		allanime_id    TEXT    NOT NULL,
		episode_number INTEGER NOT NULL,
		playback_time  INTEGER NOT NULL,
		duration       INTEGER NOT NULL,
		title          TEXT,
		last_updated   TEXT    NOT NULL,
		PRIMARY KEY (anilist_id, allanime_id)
	);`
	if _, err = db.Exec(schema); err != nil {
		db.Close()
		panic(err)
	}

	upsert, err := db.Prepare(`
		INSERT INTO anime_progress
		 (anilist_id, allanime_id, episode_number,
		  playback_time, duration, title, last_updated)
		VALUES (?,?,?,?,?,?,?)
		ON CONFLICT(anilist_id, allanime_id) DO UPDATE SET
		  episode_number = excluded.episode_number,
		  playback_time  = excluded.playback_time,
		  duration       = excluded.duration,
		  title          = excluded.title,
		  last_updated   = excluded.last_updated;`)
	if err != nil {
		db.Close()
		panic(err)
	}

	return &LocalTracker{db: db, upsertPS: upsert}
}

/*────────────────────────────────────────────────────────────────────────────*
 |  Escrita (UPSERT preparado)                                               |
 *────────────────────────────────────────────────────────────────────────────*/

func (t *LocalTracker) UpdateProgress(a Anime) error {
	a.LastUpdated = time.Now().UTC()
	_, err := t.upsertPS.Exec(
		a.AnilistID, a.AllanimeID,
		a.EpisodeNumber, a.PlaybackTime,
		a.Duration, a.Title, a.LastUpdated.Format(time.RFC3339),
	)
	return err
}

/*────────────────────────────────────────────────────────────────────────────*
 |  Leitura / deleção                                                        |
 *────────────────────────────────────────────────────────────────────────────*/

// GetAnime devolve nil se não existir.
func (t *LocalTracker) GetAnime(anilistID int, allanimeID string) (*Anime, error) {
	row := t.db.QueryRow(
		`SELECT episode_number, playback_time, duration,
		        title, last_updated
		   FROM anime_progress
		  WHERE anilist_id=? AND allanime_id=?`,
		anilistID, allanimeID,
	)

	var a Anime
	var ts string
	a.AnilistID, a.AllanimeID = anilistID, allanimeID
	if err := row.Scan(&a.EpisodeNumber, &a.PlaybackTime,
		&a.Duration, &a.Title, &ts); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	a.LastUpdated, _ = time.Parse(time.RFC3339, ts)
	return &a, nil
}

func (t *LocalTracker) GetAllAnime() ([]Anime, error) {
	rows, err := t.db.Query(
		`SELECT anilist_id, allanime_id, episode_number,
		        playback_time, duration, title, last_updated
		   FROM anime_progress`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Anime
	for rows.Next() {
		var a Anime
		var ts string
		if err = rows.Scan(&a.AnilistID, &a.AllanimeID,
			&a.EpisodeNumber, &a.PlaybackTime,
			&a.Duration, &a.Title, &ts); err != nil {
			return nil, err
		}
		a.LastUpdated, _ = time.Parse(time.RFC3339, ts)
		list = append(list, a)
	}
	return list, rows.Err()
}

func (t *LocalTracker) DeleteAnime(anilistID int, allanimeID string) error {
	_, err := t.db.Exec(
		`DELETE FROM anime_progress
		  WHERE anilist_id=? AND allanime_id=?`,
		anilistID, allanimeID,
	)
	return err
}

/*────────────────────────────────────────────────────────────────────────────*
 |  Encerrar                                                                 |
 *────────────────────────────────────────────────────────────────────────────*/

func (t *LocalTracker) Close() error {
	if t.upsertPS != nil {
		_ = t.upsertPS.Close()
	}
	return t.db.Close()
}
