package main

import (
	"encoding/json"
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	// "updateDB"
)

func ErrCheck(e error) {
	if e != nil {
		panic(e)
	}
}
func getEnv(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error Loading .env file")
	}
	return os.Getenv(key)
}

var api = slack.New(getEnv("BOT_TOKEN"))

func main() {
	h := mux.NewRouter()
	// INIT SLACK API
	signingSecret := getEnv("SLACK_SIGNING_SECRET")
	h.HandleFunc("/event-endpoint", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				fmt.Println(ev.Channel, ev.Text)
				api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
			case *slackevents.MessageEvent:
				if ev.ClientMsgID != "" {
					api.PostMessage(ev.Channel, slack.MsgOptionText("그냥 있어..", false))
				}
				// default:
			}

		}
	})
	// GET DATA
	resp, err := soup.Get("https://cs.sogang.ac.kr/front/cmsboardlist.do?siteId=cs&bbsConfigFK=1905")
	ErrCheck(err)
	doc := soup.HTMLParse(resp)
	lis := doc.Find("div", "class", "list_box").FindAll("li")
	for _, li := range lis {
		fmt.Println(li.Find("div").Find("a").Text())
		fmt.Println("https://cs.sogang.ac.kr" + li.Find("div").Find("a").Attrs()["href"])
	}
	// HTTP INIT
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to Main Page")
	})
	http.Handle("/", h)
	http.ListenAndServe(":4567", nil)

}
