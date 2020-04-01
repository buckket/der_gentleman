package main

import (
	"database/sql"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/ahmdrz/goinsta/v2"
	"github.com/buckket/der_gentleman/database"
	"github.com/buckket/der_gentleman/models"
	"github.com/buckket/der_gentleman/utils"
	"github.com/spf13/viper"
	"log"
	"net/url"
	"os"
	"time"
)

type Env struct {
	DB      *database.Database
	Insta   *goinsta.Instagram
	Twitter *anaconda.TwitterApi
	Target  int64
}

var cfgFile string

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	initConfig()

	env := Env{}

	db, err := database.New(viper.GetString("DATABASE_FILE"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	env.DB = db

	db.AutoMigrate()

	err = env.DB.CreateTableProfiles()
	if err != nil {
		log.Fatal(err)
	}

	err = env.DB.CreateTableComments()
	if err != nil {
		log.Fatal(err)
	}

	env.Twitter = anaconda.NewTwitterApiWithCredentials(viper.GetString("TWITTER_ACCESS_TOKEN"),
		viper.GetString("TWITTER_ACCESS_TOKEN_SECRET"),
		viper.GetString("TWITTER_CONSUMER_KEY"),
		viper.GetString("TWITTER_CONSUMER_SECRET"))
	_, err = env.Twitter.GetSelf(url.Values{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = os.Stat(viper.GetString("GOINSTA_FILE"))
	if err != nil {
		env.Insta = goinsta.New(viper.GetString("INSTAGRAM_USERNAME"), viper.GetString("INSTAGRAM_PASSWORD"))
		err = env.Insta.Login()
		if err != nil {
			log.Fatal(err)
		}
		env.Insta.Export(viper.GetString("GOINSTA_FILE"))
	} else {
		env.Insta, err = goinsta.Import(viper.GetString("GOINSTA_FILE"))
		if err != nil {
			log.Fatal(err)
		}
	}

	env.Target = viper.GetInt64("INSTAGRAM_TARGET")
	profile, err := env.Insta.Profiles.ByID(env.Target)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Target confirmed: %s", profile.Username)

	followers := profile.Followers()
	err = followers.Error()
	if err != nil {
		log.Fatal(err)
	}

	for followers.Next() {
		for _, u := range followers.Users {
			env.handleUser(&u)
		}
	}
}

func (env *Env) handleUser(user *goinsta.User) {
	dbProfile, err := env.DB.ProfileByIGID(user.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("error while fetching profile from database: %s", err)
			return
		}
		dbProfile = &models.Profile{
			IGID:      user.ID,
			Username:  user.Username,
			LastCheck: time.Time{},
		}
		_, err = env.DB.InsertProfile(dbProfile)
		if err != nil {
			log.Printf("error while creating database profile: %s", err)
			return
		}
	}

	if time.Since(dbProfile.LastCheck) < 1*time.Hour {
		log.Printf("Skipping %s, was already checked recently", user.Username)
		return
	} else {
		log.Printf("Checking %s", user.Username)
	}

	feed := user.Feed()
	feed.Next()
	err = feed.Error()
	if err != nil {
		return
	}

	for _, item := range feed.Items {
		env.handleItem(&item)
	}

	dbProfile.LastCheck = time.Now()
	err = env.DB.ProfileUpdate(dbProfile)
	if err != nil {
		log.Printf("error while updating profile: %s", err)
		return
	}
	time.Sleep(time.Minute)
}

func (env *Env) handleItem(item *goinsta.Item) {
	item.Comments.Sync()
	for item.Comments.Next() {
		for _, comment := range item.Comments.Items {
			if comment.UserID == env.Target {
				dbComment, err := env.DB.CommentByIGID(comment.ID)
				if err != nil && err != sql.ErrNoRows {
					log.Printf("error while fetching comment from database: %s", err)
					return
				}

				if dbComment == nil {
					dbComment = &models.Comment{
						IGID:          comment.ID,
						Text:          comment.Text,
						OpProfileIGID: item.User.ID,
						OpCode:        item.Code,
					}
					_, err := env.DB.InsertComment(dbComment)
					if err != nil {
						log.Printf("error while inserting comment to database: %s", err)
					}
					log.Printf("[%s] (%d) %s", item.User.Username, dbComment.IGID, dbComment.Text)
					tweet, err := env.Twitter.PostTweet(fmt.Sprintf("%s %s",
						utils.TruncateString(comment.Text, 256),
						dbComment.GenerateURL()), url.Values{})
					if err != nil {
						log.Fatal(err)
					}
					log.Printf("%s - %s", dbComment.GenerateURL(),
						utils.GenerateTweetURL(viper.GetString("TWITTER_USERNAME"), tweet.Id))

				} else {
					log.Printf("Old comment, skipping...")
				}
			}
		}
	}
}
