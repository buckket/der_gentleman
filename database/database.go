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
			op_profile	INTEGER NOT NULL,
			op_code		VARCHAR(255) NOT NULL,
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

func (db *Database) ProfileByIGID(ig_id int64) (*models.Profile, error) {
	profile := models.Profile{}
	err := db.QueryRow("SELECT id, ig_id, username, last_check FROM profiles WHERE ig_id = ?", ig_id).Scan(
		&profile.ID, &profile.IGID, &profile.Username, &profile.LastCheck)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (db *Database) ProfilesByLastChanged() (map[int64]time.Time, error) {
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

	res, err := tx.Exec(`INSERT INTO comments(ig_id, text, op_profile, op_code) VALUES(?, ?, ?, ?);`,
		comment.IGID, comment.Text, comment.OpProfileIGID, comment.OpCode)
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

func (db *Database) CommentByIGID(ig_id int64) (*models.Comment, error) {
	comment := models.Comment{}
	err := db.QueryRow("SELECT id, ig_id, text, op_profile, op_code FROM comments WHERE ig_id = ?", ig_id).Scan(
		&comment.ID, &comment.IGID, &comment.Text, &comment.OpProfileIGID, &comment.OpCode)
	if err != nil {
		return nil, err
	}
	return &comment, nil
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
