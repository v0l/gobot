package main

import (
	"crypto/tls"
	"fmt"

	"github.com/thoj/go-ircevent"
)

type RemoteIrc struct {
	main    *irc.Connection `json:"-"`
	i       *irc.Connection `json:"-"`
	server  string          `json:"server"`
	nick    string          `json:"nick"`
	channel string          `json:"channel"`
	ssl     bool            `json:"ssl"`
}

func (r *RemoteIrc) GetChanName() string {
	return fmt.Sprintf("#m_%v", r.channel)
}

func (r *RemoteIrc) OnPrivMsg(e *irc.Event) {
	r.main.Privmsgf(r.GetChanName(), "<%s>: %s", e.Nick, e.Message())
}

func (r *RemoteIrc) Run(server, nick, ch string, t bool) bool {
	r.server = server
	r.nick = nick
	r.channel = ch
	r.ssl = t

	return r.Start()
}

func (r *RemoteIrc) Start() bool {
	r.main.Join(r.GetChanName())
	r.main.SendRawf("TOPIC %v :%v\n", r.channel, r.server)

	r.i = irc.IRC(r.nick, r.nick)
	r.i.UseTLS = r.ssl
	r.i.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := r.i.Connect(r.server)
	if err == nil {
		r.i.AddCallback("001", func(e *irc.Event) {
			e.Connection.Join(r.channel)
		})
		r.i.AddCallback("PRIVMSG", r.OnPrivMsg)
		return true
	}

	return false
}

func (r *RemoteIrc) SendPrivmsg(msg string) {
	if r.i != nil {
		r.i.Privmsg(r.channel, msg)
	}
}

func (r *RemoteIrc) Stop() {
	if r.i != nil {
		if r.i.Connected() {
			r.i.Quit()
			r.i.Disconnect()
		}
	}
}
