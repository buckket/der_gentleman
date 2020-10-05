package cmd

import (
	"github.com/ChimeraCoder/anaconda"
	"github.com/ahmdrz/goinsta/v2"
	"github.com/buckket/der_gentleman/database"
	"github.com/buckket/der_gentleman/scraper"
	"github.com/buckket/der_gentleman/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-input"
	"log"
	"net/url"
	"os"
)

func init() {
	rootCmd.AddCommand(scrapeCmd)
}

var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Scrape IG for new data",
	Run:   scrape,
}

func scrape(cmd *cobra.Command, args []string) {
	env := scraper.Env{}

	db, err := database.New(viper.GetString("DATABASE_FILE"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	env.DB = db

	err = db.AutoMigrate()
	if err != nil {
		log.Fatal(err)
	}

	err = env.DB.CreateTableProfiles()
	if err != nil {
		log.Fatal(err)
	}

	err = env.DB.CreateTableComments()
	if err != nil {
		log.Fatal(err)
	}

	err = env.DB.CreateTableLikes()
	if err != nil {
		log.Fatal(err)
	}

	err = env.DB.CreateTableMedia()
	if err != nil {
		log.Fatal(err)
	}

	env.Limit = &utils.LimitController{}

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
		env.Limit.WaitBeforeRequest()
		err = env.Insta.Login()
		if err != nil {
			switch v := err.(type) {
			case goinsta.ChallengeError:
				env.Limit.WaitBeforeRequest()
				err := env.Insta.Challenge.Process(v.Challenge.APIPath)
				if err != nil {
					log.Fatal(err)
				}

				ui := &input.UI{
					Writer: os.Stdout,
					Reader: os.Stdin,
				}

				query := "Enter security code: "
				code, err := ui.Ask(query, &input.Options{
					Default:  "000000",
					Required: true,
					Loop:     true,
				})
				if err != nil {
					log.Fatal(err)
				}

				env.Limit.WaitBeforeRequest()
				err = env.Insta.Challenge.SendSecurityCode(code)
				if err != nil {
					log.Fatal(err)
				}

				env.Insta.Account = env.Insta.Challenge.LoggedInUser
			default:
				log.Fatal(err)
			}
		}
		env.Insta.Export(viper.GetString("GOINSTA_FILE"))
	} else {
		env.Insta, err = goinsta.Import(viper.GetString("GOINSTA_FILE"))
		if err != nil {
			log.Fatal(err)
		}
	}

	env.Limit.WaitBeforeRequest()
	env.Target, err = env.Insta.Profiles.ByID(viper.GetInt64("INSTAGRAM_TARGET"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Target confirmed: %s", env.Target.Username)

	users, err := env.GetUserIG(false)
	if err != nil {
		log.Printf("error while getting users from instagram: %s", err)
		users, err = env.GetUserDB()
		if err != nil {
			log.Fatalf("error while getting users from database: %s", err)
		}
	}

	env.Stats = &scraper.Stats{
		TotalUsers:     len(users),
		CompletedUsers: 0,
	}

	for _, u := range users {
		env.HandleUser(&u)
		env.Stats.CompletedUsers++
	}

	env.Limit.WaitBeforeRequest()
}
