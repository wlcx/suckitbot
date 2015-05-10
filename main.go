package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/docopt/docopt-go"
	ircevent "github.com/thoj/go-ircevent"
)

type Clue struct {
	Question string
	Answer   string
	Category struct {
		Title string
	}
	Value int
}

type Answer struct {
	Answer string
	Dest   string
}

func getRandomClues(n int) (clues *[]Clue, err error) {
	res, err := http.Get("http://jservice.io/api/random")
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &clues)
	if err != nil {
		return
	}
	return
}

func handlePrivmsg(event *ircevent.Event) {
	dest := event.Arguments[0]
	if dest == nick { // This is a query, set dest to querying nick
		dest = event.Nick
	}

	if event.Message() == "@clue" {
		clue <- dest
	} else if strings.HasPrefix(event.Message(), nick+": ") {
		ans <- Answer{
			strings.TrimPrefix(event.Message(), nick+": "),
			dest,
		}
	} else {
		log.Printf("msg:%s nick:%s chan:%s", event.Message(), event.Nick, event.Arguments[0])
	}

}

var nick string
var clue chan string
var ans chan Answer

func parseArgs() (args map[string]interface{}, err error) {
	usage := `suckitbot.
Usage:
  suckitbot [-hv] <#channel> [--nick=<nick>]
Options:
  -n --nick=<nick>            Bot IRC nick. [default: trebekbot]
  -h --help                   Show this help message.
  -v --version                Show version.`

	args, err = docopt.Parse(usage, nil, true, "suckitbot 0.0", false)
	return
}

func prepAnswer(answer string) string {
	r := strings.NewReplacer(
		"(", "",
		")", "",
		"'", "",
		"\"", "",
		"?", "",
	)
	answer = r.Replace(answer)
	answer = strings.TrimSpace(strings.ToLower(answer))
	for _, s := range []string{
		"what", "whats",
		"who", "whos",
		"where", "wheres",
		"is", "are", "was", "were",
		"a", "an", "the",
	} {
		answer = strings.TrimPrefix(answer, s+" ")
	}
	return answer
}

func main() {
	args, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}
	nick = args["--nick"].(string)
	irc := ircevent.IRC(nick, "Alex Trebek")
	irc.Connect("chat.freenode.net:6667")
	irc.Join(args["<#channel>"].(string))
	defer irc.Quit()

	irc.AddCallback("PRIVMSG", handlePrivmsg)

	clue = make(chan string)
	ans = make(chan Answer)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt)
	currentquestions := make(map[string]*Clue)
MainLoop:
	for {
		select {
		case destination := <-clue:
			clues, err := getRandomClues(1)
			if err != nil {
				log.Println(err)
				irc.Notice(destination, "Something went wrong there, try again soon")
			} else {
				clue := (*clues)[0]
				irc.Notice(destination, fmt.Sprintf("%s for %d: %s", clue.Category.Title, clue.Value, clue.Question))
				currentquestions[destination] = &clue
			}
		case answer := <-ans:
			preppedAns := prepAnswer(answer.Answer)
			clue := currentquestions[answer.Dest]
			isright := prepAnswer(clue.Answer) == preppedAns
			if isright {
				irc.Notice(answer.Dest, "Correct!")
			} else {
				irc.Notice(answer.Dest, fmt.Sprintf("Nope! The correct answer was: %s", clue.Answer))
			}
		case <-sigs:
			break MainLoop
		}
	}
}
