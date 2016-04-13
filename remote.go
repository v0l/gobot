package main

import (
	"crypto/tls"
	"fmt"

	"github.com/thoj/go-ircevent"
)

type RemoteIrc struct {
	Main     *irc.Connection `json:"-"`
	I        *irc.Connection `json:"-"`
	Server   string          `json:"server"`
	Nick     string          `json:"nick"`
	Channels []string        `json:"channels"`
	Ssl      bool            `json:"ssl"`
}

func (r *RemoteIrc) HasChan(ch string) bool {
	for _, v := range r.Channels {
		if v == ch {
			return true
		}
	}
	return false
}
func (r *RemoteIrc) OnPrivMsg(e *irc.Event) {
	r.Main.Privmsgf(fmt.Sprintf("#m_%v", e.Arguments[0]), "<%s>: %s", e.Nick, e.Message())
}

func (r *RemoteIrc) JoinChan(ch string) {
	r.I.Join(ch)
	r.Main.Join(fmt.Sprintf("#m_%v", ch))
	r.Main.SendRawf("TOPIC %v :%v\n", ch, r.Server)
}

func (r *RemoteIrc) Run(server, nick, ch string, t bool) bool {
	r.Server = server
	r.Nick = nick
	r.Channels = append(r.Channels, ch)
	r.Ssl = t

	return r.Start()
}

func (r *RemoteIrc) Start() bool {
	r.I = irc.IRC(r.Nick, r.Nick)
	r.I.UseTLS = r.Ssl
	r.I.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := r.I.Connect(r.Server)
	if err == nil {
		r.I.AddCallback("001", func(e *irc.Event) {
			for _, v := range r.Channels {
				r.JoinChan(v)
			}
		})
		r.I.AddCallback("PRIVMSG", r.OnPrivMsg)
		return true
	}

	return false
}

func (r *RemoteIrc) SendPrivmsg(msg string, ch string) {
	if r.I != nil {
		r.I.Privmsg(ch, msg)
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
