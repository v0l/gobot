package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/thoj/go-ircevent"
)

type TropoResponse struct {
	Success bool   `xml:"success"`
	Token   string `xml:"token"`
	ID      string `xml:"id"`
}

type Tropo struct {
}

func (*Tropo) SipCall(e *irc.Event, pn string, txt string) {
	req, re := http.Get(fmt.Sprintf("https://api.tropo.com/1.0/sessions?action=create&token=%s&numbertodial=%s&msg=%s&mode=SIP", opt.TropoToken, pn, url.QueryEscape(txt)))
	req.Header.Set("Content-Type", "application/json")
	if re == nil {
		defer req.Body.Close()
		body, _ := ioutil.ReadAll(req.Body)
		var trsp = TropoResponse{}
		je := xml.Unmarshal(body, &trsp)
		if je == nil {
			e.Connection.Privmsgf(e.Arguments[0], "[%s]: SIP call to %s started!", e.Nick, pn)
			e.Connection.Privmsgf(e.Arguments[0], "[%s]: A recording of this call will be available here: http://irc.harkin.me:6660/rec/%s.mp3", e.Nick, strings.Replace(trsp.ID, "\n", "", -1))
		} else {
			e.Connection.Privmsgf(e.Arguments[0], "[%s]: Call to %s started! (Failed to parse response: %s)", e.Nick, pn, je)
		}
	}
}

func (*Tropo) Call(e *irc.Event, pn string, txt string) {
	req, re := http.Get(fmt.Sprintf("https://api.tropo.com/1.0/sessions?action=create&token=%s&numbertodial=%s&msg=%s&mode=PTSN", opt.TropoToken, pn, url.QueryEscape(txt)))
	req.Header.Set("Content-Type", "application/json")
	if re == nil {
		body, _ := ioutil.ReadAll(req.Body)
		var trsp = TropoResponse{}
		je := xml.Unmarshal(body, &trsp)
		if je == nil {
			e.Connection.Privmsgf(e.Arguments[0], "[%s]: Call to %s started!", e.Nick, pn)
			e.Connection.Privmsgf(e.Arguments[0], "[%s]: A recording of this call will be available here: http://irc.harkin.me:6660/rec/%s.mp3", e.Nick, strings.Replace(trsp.ID, "\n", "", -1))
		} else {
			e.Connection.Privmsgf(e.Arguments[0], "[%s]: Call to %s started! (Failed to parse response: %s)", e.Nick, pn, je)
		}
	}
}

func (*Tropo) Sms(e *irc.Event, pn string, txt string) {
	_, re := http.Get(fmt.Sprintf("https://api.tropo.com/1.0/sessions?action=create&token=%s&numbertodial=%s&msg=%s&mode=SMS", opt.TropoTokenMsg, pn, url.QueryEscape(txt)))
	if re == nil {
		e.Connection.Privmsgf(e.Arguments[0], "[%s]: SMS sent to %s", e.Nick, pn)
	}
}
