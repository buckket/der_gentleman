package cmd

import (
	"fmt"
	"github.com/ahmdrz/goinsta/v2"
	"github.com/buckket/der_gentleman/database"
	"github.com/buckket/der_gentleman/scraper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

func init() {
	rootCmd.AddCommand(followingsCmd)
}

var followingsCmd = &cobra.Command{
	Use:   "followings",
	Short: "Export Followings",
	Run:   followings,
}

func followings(cmd *cobra.Command, args []string) {
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

	f, err := os.Create("followings.txt")
	if err != nil {
		log.Fatalf("error while opening followings.txt: %s", err)
	}
	defer f.Close()

	for _, u := range users {
		_, err = f.WriteString(fmt.Sprintf("%s\n", u.Username))
		if err != nil {
			log.Printf("error writting to followings.txt: %s", err)
		}
	}
}
