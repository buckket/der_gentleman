package cmd

import (
	"encoding/json"
	"github.com/ChimeraCoder/anaconda"
	"github.com/buckket/der_gentleman/utils"
	"github.com/mb-14/gomarkov"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"strings"
)

func init() {
	rootCmd.AddCommand(botCmd)
}

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Reply with generated data",
	Run:   bot,
}

func bot(cmd *cobra.Command, args []string) {
	var chain gomarkov.Chain
	data, err := ioutil.ReadFile("model.json")
	if err != nil {
		log.Fatalf("could not load model.json: %s", err)
	}
	err = json.Unmarshal(data, &chain)
	if err != nil {
		log.Fatalf("could not load markov chain: %s", err)
	}

	twitter := anaconda.NewTwitterApiWithCredentials(viper.GetString("TWITTER_ACCESS_TOKEN"),
		viper.GetString("TWITTER_ACCESS_TOKEN_SECRET"),
		viper.GetString("TWITTER_CONSUMER_KEY"),
		viper.GetString("TWITTER_CONSUMER_SECRET"))
	_, err = twitter.GetSelf(url.Values{})
	if err != nil {
		log.Fatal(err)
	}

	sinceID := readSinceID()
	log.Printf("Using sinceID: %d", sinceID)

	v := url.Values{}
	v.Add("since_id", strconv.FormatInt(sinceID, 10))
	tl, err := twitter.GetMentionsTimeline(v)
	if err != nil {
		log.Printf("could not get mentions: %s", err)
	}

	for _, m := range tl {
		v := url.Values{}
		v.Add("auto_populate_reply_metadata", "true")
		v.Add("in_reply_to_status_id", strconv.FormatInt(m.Id, 10))
		_, err := twitter.Favorite(m.Id)
		if err != nil {
			log.Printf("could not fav tweet: %s", err)
		}
		t, err := twitter.PostTweet(utils.TruncateString(generateTweet(&chain), 280), v)
		if err != nil {
			log.Printf("could not tweet: %s", err)
		}
		log.Printf("New tweet in reply to %s", t.InReplyToScreenName)
		if m.Id > sinceID {
			sinceID = m.Id
			writeSinceID(sinceID)
		}
	}
}

func generateTweet(chain *gomarkov.Chain) string {
	tokens := []string{gomarkov.StartToken}
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := chain.Generate(tokens[(len(tokens) - 1):])
		tokens = append(tokens, next)
	}
	return strings.Join(tokens[1:len(tokens)-1], " ")
}

func readSinceID() int64 {
	idFile, err := ioutil.ReadFile("bot_since_id")
	if err != nil {
		log.Print(err)
		return 1
	}
	id, err := strconv.Atoi(strings.TrimSpace(string(idFile)))
	if err != nil {
		log.Print(err)
		return 1
	}
	return int64(id)
}

func writeSinceID(id int64) {
	err := ioutil.WriteFile("bot_since_id", []byte(strconv.FormatInt(id, 10)), 0644)
	if err != nil {
		log.Print(err)
	}
}
