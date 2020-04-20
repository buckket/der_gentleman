package models

import (
	"fmt"
	"time"
)

type Profile struct {
	ID        int64
	IGID      int64
	Username  string
	LastCheck time.Time
}

type Comment struct {
	ID            int64
	IGID          int64
	Text          string
	CreatedAt     time.Time
	OpProfileIGID int64
	OpMediaIGID   string
}

type Like struct {
	ID            int64
	CreatedAt     time.Time
	OpProfileIGID int64
	OpMediaIGID   string
}

type Media struct {
	ID            int64
	IGID          string
	CreatedAt     time.Time
	OpProfileIGID int64
	OpCode        string
	Likes         int
	Comments      int
}

func (m *Media) GenerateURL() string {
	return fmt.Sprintf("https://www.instagram.com/p/%s/", m.OpCode)
}
