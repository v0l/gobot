package main

import (
	"math/rand"
	"strings"

	"github.com/Lazin/go-ngram"
	"github.com/kennygrant/sanitize"
	"github.com/thoj/go-ircevent"
)

type NChat struct {
	Index *ngram.NGramIndex
	grams int
}

func (n *NChat) Init(grams int) {
	n.grams = grams
	idx, _ := ngram.NewNGramIndex(ngram.SetN(grams))
	n.Index = idx
}

func (n *NChat) Chat(msg string) {
	if strings.Index(msg, " ") > 0 {
		//remove retweet crap
		if strings.Index(msg, "RT") == 0 {
			msg = msg[strings.Index(msg, ":"):]
		}

		msg = sanitize.Accents(msg)
		msg = strings.Replace(msg, "\n", "", -1)
		ms := strings.Split(msg, " ")
		for i := 0; i < len(ms); i++ {
			if i+n.grams <= len(ms) {
				n.Index.Add(strings.Join(ms[i:i+n.grams], " "))
			} else {
				n.Index.Add(strings.Join(ms[i:len(ms)], " "))
			}
		}
	}
}

func (n *NChat) Next(seed string) string {
	//pick a random length up to 50 words
	l := rand.Intn(50)
	var ret string
	var last string

	for i := 0; i < l/n.grams; i++ {
		if i == 0 {
			res, re := n.Index.BestMatch(seed)
			if re == nil {
				wd, wde := n.Index.GetString(res.TokenID)
				if wde == nil && last != wd {
					ret = ret + " " + wd
					last = wd
				}
			}
		} else {
			idx := strings.LastIndex(ret, " ")
			if idx == -1 {
				idx = 0
			}
			res, re := n.Index.BestMatch(ret[idx:])
			if re == nil {
				wd, wde := n.Index.GetString(res.TokenID)
				if wde == nil && last != wd {
					//also look 1 ahead
					idx1 := strings.LastIndex(wd, " ")
					if idx1 == -1 {
						idx1 = 0
					}
					res1, re1 := n.Index.BestMatch(wd[idx1:])
					if re1 != nil {
						ret = ret + " " + wd
						last = wd
					} else if res1 != res {
						ret = ret + " " + wd
						last = wd
					}
				}
			}
		}
	}

	if len(ret) > 0 {
		return "\x03" + "9" + ret + "\x0F"
	} else {
		return ""
	}
}

func (n *NChat) Info(e *irc.Event) {

}
