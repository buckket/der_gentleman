package cmd

import (
	"encoding/json"
	"github.com/buckket/der_gentleman/database"
	"github.com/mb-14/gomarkov"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"strings"
)

func init() {
	rootCmd.AddCommand(chainsCmd)
}

var chainsCmd = &cobra.Command{
	Use:   "chains",
	Short: "Generate Markov chains",
	Run:   chains,
}

func chains(cmd *cobra.Command, args []string) {
	db, err := database.New(viper.GetString("DATABASE_FILE"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	texts, err := db.Comments()
	if err != nil {
		log.Fatalf("could not get comments: %s", err)
	}

	chain := gomarkov.NewChain(2)
	for _, s := range *texts {
		chain.Add(strings.Split(s, " "))
	}

	jsonObj, _ := json.Marshal(chain)
	err = ioutil.WriteFile("model.json", jsonObj, 0644)
	if err != nil {
		log.Fatalf("could not write mode.json: %s", err)
	}
}
