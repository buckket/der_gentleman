package scraper

import (
	"database/sql"
	"github.com/ahmdrz/goinsta/v2"
	"github.com/buckket/der_gentleman/models"
	"log"
	"sort"
	"time"
)

func (env *Env) GetUserIG(sync bool) ([]goinsta.User, error) {
	var users []goinsta.User

	followers := env.Target.Following()
	for followers.Next() {
		err := followers.Error()
		if err != nil && err != goinsta.ErrNoMore {
			return nil, err
		}
		users = append(users, followers.Users...)
	}

	err := followers.Error()
	if err != nil && err != goinsta.ErrNoMore {
		return nil, err
	}

	lastChangedMap, err := env.DB.ProfilesMapByLastChanged()
	if err != nil {
		return nil, err
	}

	sort.Slice(users, func(i, j int) bool {
		dbI, ok := lastChangedMap[users[i].ID]
		if !ok {
			return true
		}
		dbJ, ok := lastChangedMap[users[j].ID]
		if !ok {
			return false
		}
		return dbI.Before(dbJ)
	})

	if sync {
		for _, u := range users {
			if u.IsPrivate {
				err := u.Follow()
				if err != nil {
					log.Printf("coult not follow %s: %s", u.Username, err)
				}
				log.Printf("send follow request to %s", u.Username)
				time.Sleep(1 * time.Minute)
			}
		}
	}

	return users, nil
}

func (env *Env) GetUserDB() ([]goinsta.User, error) {
	var users []goinsta.User

	followers, err := env.DB.ProfilesSorted()
	if err != nil {
		return nil, err
	}

	for _, f := range followers {
		u := goinsta.User{
			ID:       f.IGID,
			Username: f.Username,
		}
		u.SetInstagram(env.Insta)
		users = append(users, u)
	}
	return users, nil
}

func (env *Env) HandleUser(user *goinsta.User) {
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

	if user.Username != dbProfile.Username {
		log.Printf("Username changed from %s to %s, updating database", dbProfile.Username, user.Username)
		dbProfile.Username = user.Username
		err = env.DB.ProfileUpdate(dbProfile)
		if err != nil {
			log.Printf("error while updating profile: %s", err)
			return
		}
	}

	if time.Since(dbProfile.LastCheck) < 1*time.Hour {
		log.Printf("Skipping %s, was already checked recently", user.Username)
		return
	}

	feed := user.Feed()
	feed.Next()
	err = feed.Error()
	if err != nil && err != goinsta.ErrNoMore {
		log.Printf("error while fetching user feed: %s", err)
		return
	}

	log.Printf("[%d/%d] Checking %s (%d items)", env.Stats.CompletedUsers, env.Stats.TotalUsers, user.Username, len(feed.Items))

	for _, item := range feed.Items {
		env.handleItem(&item)
	}

	dbProfile.LastCheck = time.Now()
	err = env.DB.ProfileUpdate(dbProfile)
	if err != nil {
		log.Printf("error while updating profile: %s", err)
		return
	}
}
