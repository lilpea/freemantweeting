package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain     map[string][]string
	prefixLen int
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen}
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)
	p := make(Prefix, c.prefixLen)
	i := 0
	var starts []int
	var words []string
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}
		if unicode.IsUpper([]rune(s)[0]) {
			starts = append(starts, i)
		}
		words = append(words, s)
		i++
	}
	words = append(words[starts[rand.Intn(len(starts))]:], words[:starts[rand.Intn(len(starts))]]...)
	for i = 0; i < len(words); {
		s := words[i]
		key := p.String()
		c.chain[key] = append(c.chain[key], s)
		p.Shift(s)
		i++
	}
}

// Generate returns a string of at most n words generated from Chain.
func (c *Chain) Generate(n int) string {
	p := make(Prefix, c.prefixLen)
	var words []string
	keepgenerating := true
	for keepgenerating {
		choices := c.chain[p.String()]
		next := choices[rand.Intn(len(choices))]
		words = append(words, next)
		punct, err := regexp.MatchString(".*[.?!]$", next)
		if err != nil {
			log.Fatalf("error in matching regex: %v", err)
		}
		if len(words) >= n && punct {
			keepgenerating = false
		}
		p.Shift(next)
	}
	return strings.Join(words, " ")
}

func main() {
	baseText, err := os.Open("data.txt")
	if err != nil {
		log.Fatalf("error in opening data.txt: %v", err)
	}
	jsonConf, err := ioutil.ReadFile("configuration.json")
	if err != nil {
		log.Fatalf("error in opening configuration.json: %v", err)
	}
	jsonAuth, err := ioutil.ReadFile("authentication.json")
	if err != nil {
		log.Fatalf("error in opening authentication.json: %v", err)
	}
	var conf map[string]int
	var auth map[string]string
	err = json.Unmarshal(jsonConf, &conf)
	if err != nil {
		log.Fatalf("error in unmarshalling configuration.json: %v", err)
	}
	err = json.Unmarshal(jsonAuth, &auth)
	if err != nil {
		log.Fatalf("error in unmarshalling authentication.json: %v", err)
	}

	rand.Seed(time.Now().UnixNano())   // Seed the random number generator.
	c := NewChain(conf["prefixCount"]) // Initialize a new Chain.
	c.Build(baseText)                  // Build the Markov Chain
	var text string

	//Somewhat hacky solution for meeting twitter's character limits,
	//remove this section if this doesn't matter for you
	keepgenerating := true
	for keepgenerating {
		text = c.Generate(conf["wordCount"])
		sentencematches := regexp.MustCompile("[.?!] ").FindAllStringSubmatchIndex(text, -1)
		for i := 0; i < len(sentencematches); {
			if sentencematches[i][0]-1 >= conf["charCount"]-10 &&
				sentencematches[i][0]-1 <= conf["charCount"] {
				text = text[:sentencematches[i][0]+1]
				keepgenerating = false
				break
			}
			i++
		}
	}

	//If you want for this to be a CLI command, then you can replace this section with "fmt.Println(text)"
	config := oauth1.NewConfig(auth["consumerKey"], auth["consumerSecret"])
	token := oauth1.NewToken(auth["accessToken"], auth["accessSecret"])
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	_, _, err = client.Statuses.Update(text, nil)
	if err != nil {
		log.Fatalf("error in sending tweet: %v", err)
	}
}
