package scraper

import (
	"database/sql"
	"fmt"
	"github.com/ahmdrz/goinsta/v2"
	"github.com/buckket/der_gentleman/models"
	"github.com/buckket/der_gentleman/utils"
	"github.com/spf13/viper"
	"log"
	"net/url"
	"time"
)

func (env *Env) handleComments(item *goinsta.Item, media *models.Media) {
	if item.CommentsDisabled {
		log.Printf("Comments disabled, skipping...")
		return
	}

	if item.CommentCount > 1000 {
		log.Printf("Too many comments, skipping...")
		return
	}

	item.Comments.Sync()
	for item.Comments.Next() {
		for _, comment := range item.Comments.Items {
			if comment.UserID != env.Target.ID {
				continue
			}
			dbComment, err := env.DB.CommentByIGID(comment.ID)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("error while fetching comment from database: %s", err)
				return
			}
			if dbComment == nil {
				dbComment = &models.Comment{
					IGID:          comment.ID,
					Text:          comment.Text,
					CreatedAt:     time.Unix(comment.CreatedAtUtc, 0),
					OpProfileIGID: item.User.ID,
					OpMediaIGID:   item.ID,
				}
				_, err := env.DB.InsertComment(dbComment)
				if err != nil {
					log.Printf("error while inserting comment to database: %s", err)
					return
				}
				log.Printf("New comment: [%s] (%s) %s", item.User.Username, media.OpCode, dbComment.Text)

				if viper.GetBool("TWITTER_ENABLE") {
					td := time.Now().Sub(lastTweet)
					if td < 5*time.Minute {
						log.Printf("Last tweet was only %f minutes ago, waiting...", td.Minutes())
						time.Sleep(5*time.Minute - td)
					}

					tweet, err := env.Twitter.PostTweet(fmt.Sprintf("%s %s",
						utils.TruncateString(comment.Text, 256),
						media.GenerateURL()), url.Values{})
					if err != nil {
						log.Fatal(err)
					}

					lastTweet = time.Now()
					log.Printf("%s - %s", media.GenerateURL(),
						utils.GenerateTweetURL(viper.GetString("TWITTER_USERNAME"), tweet.Id))
				}
			} else {
				if dbComment.Text != comment.Text || dbComment.CreatedAt.UTC() != time.Unix(comment.CreatedAtUtc, 0).UTC() {
					log.Printf("Old comment but database out-of-date, updating...")
					log.Printf("Old text: %s", dbComment.Text)
					log.Printf("New text: %s", comment.Text)
					log.Printf("Old time: %s", dbComment.CreatedAt)
					log.Printf("New time: %s", time.Unix(comment.CreatedAtUtc, 0))
					dbComment.Text = comment.Text
					dbComment.CreatedAt = time.Unix(comment.CreatedAtUtc, 0)
					err = env.DB.CommentUpdate(dbComment)
					if err != nil {
						log.Printf("error while updating comment: %s", err)
						return
					}
				}
			}
		}
	}
}
