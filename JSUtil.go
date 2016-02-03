package main

import (
	"errors"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/thoj/go-ircevent"
)

type JSUtil struct {
}

func (*JSUtil) RunJS(e *irc.Event) {
	var halt = errors.New("Stahp")

	defer func() {
		if caught := recover(); caught != nil {
			if caught == halt {
				e.Connection.Privmsgf(e.Arguments[0], "!! WTF STAHP %s !!", e.Nick)
				return
			}
		}
	}()

	q := strings.TrimSpace(strings.Replace(e.Arguments[1], "!js ", "", -1))
	vm := otto.New()
	vm.Interrupt = make(chan func(), 1) // The buffer prevents blocking
	go func() {
		time.Sleep(10 * time.Second) // Stop after two seconds
		vm.Interrupt <- func() {
			panic(halt)
		}
	}()

	script, ser := vm.Compile("", q)
	if ser == nil {
		outjs, er := vm.Run(script)
		if er == nil {
			e.Connection.Privmsgf(e.Arguments[0], "[%s]: %s", e.Nick, outjs)
		} else {
			e.Connection.Privmsgf(e.Arguments[0], "Error in JS %s: %s", e.Nick, er)
		}
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "Compile error %s: %s", e.Nick, ser)
	}
}
