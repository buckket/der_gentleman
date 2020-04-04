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
	OpProfileIGID int64
	OpCode        string
	Tweeted       bool
}

func (c *Comment) GenerateURL() string {
	return fmt.Sprintf("https://www.instagram.com/p/%s/", c.OpCode)
}
