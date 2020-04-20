package database

import (
	"database/sql"
	"fmt"
	"github.com/buckket/der_gentleman/models"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Database struct {
	*sql.DB
}

func (db *Database) AutoMigrate() error {
	var version int
	err := db.QueryRow(`PRAGMA user_version`).Scan(&version)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) CreateTableProfiles() error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS profiles(
			id			INTEGER PRIMARY KEY AUTOINCREMENT,
			ig_id		INTEGER NOT NULL UNIQUE,
			username	VARCHAR(255) NOT NULL,
			last_check	DATETIME
		);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func (db *Database) CreateTableComments() error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS comments(
			id 			INTEGER PRIMARY KEY AUTOINCREMENT,
			ig_id		INTEGER NOT NULL UNIQUE,
			text 		TEXT,
			created_at	DATETIME,
			op_profile	INTEGER NOT NULL,
			op_media	VARCHAR(255) NOT NULL,
			FOREIGN KEY(op_profile) REFERENCES profiles(ig_id)
			FOREIGN KEY(op_media) REFERENCES media(ig_id)
		);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func (db *Database) CreateTableLikes() error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS likes(
			id 			INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at	DATETIME,
			op_profile	INTEGER NOT NULL,
			op_media	VARCHAR(255) NOT NULL,
			FOREIGN KEY(op_profile) REFERENCES profiles(ig_id)
			FOREIGN KEY(op_media) REFERENCES media(ig_id)
		);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func (db *Database) CreateTableMedia() error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS media(
			id 			INTEGER PRIMARY KEY AUTOINCREMENT,
			ig_id		VARCHAR(255) NOT NULL UNIQUE,
			created_at	DATETIME,
			code		VARCHAR(255) NOT NULL,
			op_profile	INTEGER NOT NULL,
			likes		INTEGER NOT NULL,
			comments	INTEGER NOT NULL,
			FOREIGN KEY(op_profile) REFERENCES profiles(ig_id)
		);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func (db *Database) InsertProfile(profile *models.Profile) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO profiles(ig_id, username, last_check) VALUES(?, ?, ?);`,
		profile.IGID, profile.Username, profile.LastCheck)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return lastID, nil
}

func (db *Database) ProfileByIGID(igID int64) (*models.Profile, error) {
	profile := models.Profile{}
	err := db.QueryRow("SELECT id, ig_id, username, last_check FROM profiles WHERE ig_id = ?", igID).Scan(
		&profile.ID, &profile.IGID, &profile.Username, &profile.LastCheck)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (db *Database) ProfilesMapByLastChanged() (map[int64]time.Time, error) {
	m := make(map[int64]time.Time)

	rows, err := db.Query("SELECT ig_id, last_check FROM profiles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var igID int64
		var lastChanged time.Time
		err = rows.Scan(&igID, &lastChanged)
		if err != nil {
			return nil, err
		}
		m[igID] = lastChanged
	}
	return m, nil
}

func (db *Database) ProfilesSorted() ([]models.Profile, error) {
	var profiles []models.Profile

	rows, err := db.Query("SELECT id, ig_id, username, last_check FROM profiles ORDER BY last_check ASC;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := models.Profile{}
		err = rows.Scan(&p.ID, &p.IGID, &p.Username, &p.LastCheck)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (db *Database) ProfileUpdate(profile *models.Profile) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE profiles SET username = ?, last_check = ? WHERE ig_id = ?`,
		profile.Username, profile.LastCheck, profile.IGID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count <= 0 {
		return fmt.Errorf("no rows affected")
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *Database) InsertComment(comment *models.Comment) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO comments(ig_id, text, op_profile, op_media, created_at) VALUES(?, ?, ?, ?, ?);`,
		comment.IGID, comment.Text, comment.OpProfileIGID, comment.OpMediaIGID, comment.CreatedAt)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return lastID, nil
}

func (db *Database) Comments() (*[]string, error) {
	var texts []string

	rows, err := db.Query("SELECT text FROM comments;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		err = rows.Scan(&text)
		if err != nil {
			return nil, err
		}
		texts = append(texts, text)
	}
	return &texts, nil
}

func (db *Database) CommentByIGID(igID int64) (*models.Comment, error) {
	comment := models.Comment{}
	err := db.QueryRow("SELECT id, ig_id, text, op_profile, op_media, created_at FROM comments WHERE ig_id = ?", igID).Scan(
		&comment.ID, &comment.IGID, &comment.Text, &comment.OpProfileIGID, &comment.OpMediaIGID, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (db *Database) CommentUpdate(comment *models.Comment) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE comments SET text = ?, created_at = ? WHERE ig_id = ?`,
		comment.Text, comment.CreatedAt, comment.IGID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count <= 0 {
		return fmt.Errorf("no rows affected")
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *Database) InsertLike(like *models.Like) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO likes(op_profile, op_media, created_at) VALUES(?, ?, ?);`,
		like.OpProfileIGID, like.OpMediaIGID, like.CreatedAt)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return lastID, nil
}

func (db *Database) LikeByIGID(igID string) (*models.Like, error) {
	like := models.Like{}
	err := db.QueryRow("SELECT id, op_profile, op_media, created_at FROM likes WHERE op_media = ?", igID).Scan(
		&like.ID, &like.OpProfileIGID, &like.OpMediaIGID, &like.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (db *Database) InsertMedia(media *models.Media) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO media(ig_id, op_profile, created_at, code, likes, comments) VALUES(?, ?, ?, ?, ?, ?);`,
		media.IGID, media.OpProfileIGID, media.CreatedAt, media.OpCode, media.Likes, media.Comments)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return lastID, nil
}

func (db *Database) MediaByIGID(igID string) (*models.Media, error) {
	media := models.Media{}
	err := db.QueryRow("SELECT id, ig_id, op_profile, created_at, code, likes, comments FROM media WHERE ig_id = ?", igID).Scan(
		&media.ID, &media.IGID, &media.OpProfileIGID, &media.CreatedAt, &media.OpCode, &media.Likes, &media.Comments)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (db *Database) MediaUpdate(media *models.Media) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE media SET code = ?, likes = ?, comments = ? WHERE ig_id = ?`,
		media.OpCode, media.Likes, media.Comments, media.IGID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count <= 0 {
		return fmt.Errorf("no rows affected")
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func New(target string) (*Database, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?&_fk=true", target))
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &Database{db}, nil
}
