package main

import (
	"crypto/tls"
	"fmt"

	"github.com/thoj/go-ircevent"
)

type RemoteIrc struct {
	Main    *irc.Connection `json:"-"`
	I       *irc.Connection `json:"-"`
	Server  string          `json:"server"`
	Nick    string          `json:"nick"`
	Channel string          `json:"channel"`
	Ssl     bool            `json:"ssl"`
}

func (r *RemoteIrc) GetChanName() string {
	return fmt.Sprintf("#m_%v", r.Channel)
}

func (r *RemoteIrc) OnPrivMsg(e *irc.Event) {
	r.Main.Privmsgf(r.GetChanName(), "<%s>: %s", e.Nick, e.Message())
}

func (r *RemoteIrc) Run(server, nick, ch string, t bool) bool {
	r.Server = server
	r.Nick = nick
	r.Channel = ch
	r.Ssl = t

	return r.Start()
}

func (r *RemoteIrc) Start() bool {
	r.Main.Join(r.GetChanName())
	r.Main.SendRawf("TOPIC %v :%v\n", r.Channel, r.Server)

	r.I = irc.IRC(r.Nick, r.Nick)
	r.I.UseTLS = r.Ssl
	r.I.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := r.I.Connect(r.Server)
	if err == nil {
		r.I.AddCallback("001", func(e *irc.Event) {
			e.Connection.Join(r.Channel)
		})
		r.I.AddCallback("PRIVMSG", r.OnPrivMsg)
		return true
	}

	return false
}

func (r *RemoteIrc) SendPrivmsg(msg string) {
	if r.I != nil {
		r.I.Privmsg(r.Channel, msg)
	}
}

func (r *RemoteIrc) Stop() {
	if r.I != nil {
		if r.I.Connected() {
			r.I.Quit()
			r.I.Disconnect()
		}
	}
}
