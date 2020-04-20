package scraper

import (
	"database/sql"
	"github.com/ahmdrz/goinsta/v2"
	"github.com/buckket/der_gentleman/models"
	"github.com/buckket/der_gentleman/utils"
	"log"
	"time"
)

func (env *Env) handleLikes(item *goinsta.Item, media *models.Media) {
	for _, topliker := range utils.TopLikers(item.Toplikers) {
		if topliker != env.Target.Username {
			continue
		}
		dbLike, err := env.DB.LikeByIGID(item.ID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("error while fetching like from database: %s", err)
			return
		}
		if dbLike == nil {
			dbLike = &models.Like{
				CreatedAt:     time.Now(),
				OpProfileIGID: item.User.ID,
				OpMediaIGID:   item.ID,
			}
			_, err := env.DB.InsertLike(dbLike)
			if err != nil {
				log.Printf("error while creating like: %s", err)
				return
			}
			log.Printf("New like: [%s] (%s)", item.User.Username, media.OpCode)
		}
		break
	}
}
