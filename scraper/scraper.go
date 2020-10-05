package scraper

import (
	"github.com/ChimeraCoder/anaconda"
	"github.com/ahmdrz/goinsta/v2"
	"github.com/buckket/der_gentleman/database"
	"github.com/buckket/der_gentleman/utils"
	"time"
)

type Stats struct {
	TotalUsers     int
	CompletedUsers int
}

type Env struct {
	DB      *database.Database
	Insta   *goinsta.Instagram
	Twitter *anaconda.TwitterApi
	Target  *goinsta.User
	Stats   *Stats
	Limit   *utils.LimitController
}

var lastTweet time.Time
