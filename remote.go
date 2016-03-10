package main

import (
	"crypto/tls"

	"github.com/thoj/go-ircevent"
)

type RemoteIrc struct {
	main    *irc.Connection
	i       *irc.Connection
	server  string
	nick    string
	channel string
}

func (r *RemoteIrc) OnPrivMsg(e *irc.Event) {
	r.main.Privmsgf("#lobby", "[%s] %s", e.Nick, e.Message())
}

func (r *RemoteIrc) Run(server, nick, ch string, t bool) {
	r.server = server
	r.nick = nick
	r.channel = ch

	r.i = irc.IRC(nick, nick)
	r.i.UseTLS = t
	r.i.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := r.i.Connect(server)
	if err == nil {
		r.i.AddCallback("001", func(e *irc.Event) {
			e.Connection.Join(ch)
		})
		r.i.AddCallback("PRIVMSG", r.OnPrivMsg)
	}

	r.i.Loop()
}

func (r *RemoteIrc) SendPrivmsg(ch, msg string) {
	r.i.Privmsg(ch, msg)
}

func (r *RemoteIrc) Stop() {
	r.i.Disconnect()
}