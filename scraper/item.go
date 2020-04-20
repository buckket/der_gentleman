package scraper

import (
	"database/sql"
	"github.com/ahmdrz/goinsta/v2"
	"github.com/buckket/der_gentleman/models"
	"log"
	"time"
)

func (env *Env) handleItem(item *goinsta.Item) {
	dbMedia, err := env.DB.MediaByIGID(item.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error while fetching media from database: %s", err)
		return
	}
	if dbMedia == nil {
		dbMedia = &models.Media{
			IGID:          item.ID,
			CreatedAt:     time.Unix(item.TakenAt, 0),
			OpProfileIGID: item.User.ID,
			OpCode:        item.Code,
			Likes:         0,
			Comments:      0,
		}
		_, err := env.DB.InsertMedia(dbMedia)
		if err != nil {
			log.Printf("error while creating media: %s", err)
			return
		}
	}

	if item.Likes != dbMedia.Likes {
		env.handleLikes(item, dbMedia)
	}

	if item.CommentCount != dbMedia.Comments {
		env.handleComments(item, dbMedia)
	}

	if dbMedia.Likes != item.Likes || dbMedia.Comments != item.CommentCount || dbMedia.OpCode != item.Code {
		dbMedia.Likes = item.Likes
		dbMedia.Comments = item.CommentCount
		dbMedia.OpCode = item.Code
		err = env.DB.MediaUpdate(dbMedia)
		if err != nil {
			log.Printf("error while updating media: %s", err)
		}
	}
}
