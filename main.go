package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/thoj/go-ircevent"
)

type Options struct {
	Server       string
	Nick         string
	useTLS       bool
	DefaultChans []string
	Oper         bool
	OperDetails  []string
	RawPW        string

	//Youtube
	YoutubeApiKey string

	//Twitter
	TwitterAppKey    string
	TwitterAppSecret string
	TwitterTokenDir  string

	//Temp twitter vars
	TwitterAuthKey    string
	TwitterAuthSecret string
	TwitterHandle     string

	//Tropo
	TropoToken    string
	TropoTokenMsg string
	CallDir       string
}

var opt = Options{}
var twu = TwitterUtil{0, TweetLocation{}}
var rc = []RemoteIrc{}

func OnJoin(e *irc.Event) {
	if e.Nick != opt.Nick {
		e.Connection.Privmsgf(e.Arguments[0], "Welcome %s!", e.Nick)
	}
}

func Hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func OnPrivMsg(e *irc.Event) {
	if strings.Index(e.Nick, "T10") >= 0 {
		return
	}

	msg := e.Message()
	args := e.Arguments
	l := e.Connection.Log

	if strings.Index(msg, "!") == 0 {
		cmd := strings.Split(args[1], " ")
		l.Printf("CMD (%s) from: %s (%s)\n", cmd[0], e.Nick, args)

		switch strings.ToLower(cmd[0]) {
		case "!help":
			{
				e.Connection.Privmsgf(args[0], "Some help info sir.")
				e.Connection.Privmsgf(args[0], fmt.Sprintf("Twatter:\thttps://twitter.com/%s", opt.TwitterHandle))
				e.Connection.Privmsgf(args[0], "!help \t\t- Shows this message.")
				e.Connection.Privmsgf(args[0], "!ping \t\t- Get told where to go.")
				e.Connection.Privmsgf(args[0], "!raw <password> <command> \t- Run a raw IRC command from me.")
				e.Connection.Privmsgf(args[0], "!excuse (!lie)\t\t- Get out of jail free cards are dispensed here")
				e.Connection.Privmsgf(args[0], "!love \t\t- Get some coding love!")
				e.Connection.Privmsgf(args[0], "!sip <addr> <msg> \t- Calls a sip phone with a message or sound file (or both)")
				e.Connection.Privmsgf(args[0], "!call <no> <msg> \t- Calls a phone number and says a message or plays a sound eg. !call +44XX1234567 I love you")
				e.Connection.Privmsgf(args[0], "!sms <no> <msg> \t- Send an SMS to the supplied phone number")
				e.Connection.Privmsgf(args[0], "!ud (!w) <word>\t- Looks words up on urban dictionary")
				e.Connection.Privmsgf(args[0], "!s <thing>\t- Does a quick Google")
				e.Connection.Privmsgf(args[0], "!js <code> \t- Runs some JS code")
				e.Connection.Privmsgf(args[0], "!remote <server> <nick> <chan> <ssl>\t- Connectes to another IRC server and pipes chat to #lobby")
				e.Connection.Privmsgf(args[0], "!rclose <server#>\t- Close connection to remote")
				e.Connection.Privmsgf(args[0], "!thelp \t- Gets twitter command list")
				break
			}
		case "!remote":
			{
				if len(cmd) > 4 {
					nc := RemoteIrc{}
					nc.main = e.Connection
					ncs := false
					if strings.ToLower(cmd[4]) == "true" {
						ncs = true
					}

					if nc.Run(cmd[1], cmd[2], cmd[3], ncs) {
						rc = append(rc, nc)
						e.Connection.Privmsgf(args[0], "[%v] Connected to %s", e.Nick, cmd[1])
					}
				} else {
					e.Connection.Privmsgf(args[0], "Remote connections:")
					for k, v := range rc {
						e.Connection.Privmsgf(args[0], "%v: %s %s (%s)", k, v.server, v.channel, v.nick)
					}
				}
				break
			}
		case "!rclose":
			{
				if len(cmd) > 1 {
					srv, ser := strconv.Atoi(cmd[1])
					if ser == nil {
						if srv < len(rc) {
							tc := rc[srv]
							if len(rc) == 1 {
								rc = rc[:0]
							} else {
								rc = append(rc[:srv], rc[srv+1:]...)
							}

							defer tc.Stop()
						}
					}
				}
				break
			}
		case "!thelp":
			{
				twu.GetHelp(e)
				break
			}
		case "!ping":
			{
				e.Connection.Privmsgf(args[0], "STFU %s!", e.Nick)
				break
			}
		case "!raw":
			{
				pw := cmd[1]
				hpw := Hash(pw)
				if hpw == opt.RawPW {
					q := strings.TrimSpace(strings.Replace(args[1], "!raw "+pw+" ", "", -1))
					l.Printf("Running raw command %s", q)
					e.Connection.SendRawf("%s", q)
				} else {
					l.Printf("Hash of %s did not match %s", hpw, opt.RawPW)
				}
				break
			}
		case "!excuse", "!lie":
			{
				utl := new(HttpUtils)
				utl.GetExcuse(e)
				break
			}
		case "!love":
			{
				utl := new(HttpUtils)
				utl.GetLove(e)
				break
			}
		case "!s":
			{
				q := strings.TrimSpace(strings.Replace(args[1], "!s ", "", -1))
				utl := new(HttpUtils)
				utl.SearchGoogle(e, q)
				break
			}
		case "!sip":
			{
				pn := cmd[1]
				q := strings.TrimSpace(strings.Replace(args[1], "!call "+pn+" ", "", -1))

				tr := new(Tropo)
				tr.SipCall(e, pn, q)
				break
			}
		case "!call":
			{
				pn := cmd[1]
				q := strings.TrimSpace(strings.Replace(args[1], "!call "+pn+" ", "", -1))

				tr := new(Tropo)
				tr.Call(e, pn, q)
				break
			}
		case "!sms":
			{
				pn := cmd[1]
				q := strings.TrimSpace(strings.Replace(args[1], "!sms "+pn+" ", "", -1))

				tr := new(Tropo)
				tr.Sms(e, pn, q)
				break
			}
		case "!ud", "!w":
			{
				q := strings.TrimSpace(strings.Replace(strings.Replace(args[1], "!ud ", "", -1), "!w ", "", -1))

				utl := new(HttpUtils)
				utl.SearchUrbanDictionary(e, q)
				break
			}
		case "!funfact":
			{
				utl := new(HttpUtils)
				utl.GetFunFact(e)
				break
			}
		case "!js":
			{
				jsu := new(JSUtil)
				jsu.RunJS(e)
				break
			}
		case "!tf":
			{
				twu.Follow(e, cmd[1])
				break
			}
		case "!t":
			{
				q := strings.TrimSpace(strings.Replace(args[1], "!t ", "", -1))

				twu.SendTweet(e, q)
				break
			}
		case "!tr":
			{
				tid := cmd[1]
				q := strings.TrimSpace(strings.Replace(args[1], "!tr "+tid, "", -1))

				twu.SendTweetResponse(e, q, tid)
				break
			}
		case "!tdm":
			{
				usr := cmd[1]
				q := strings.TrimSpace(strings.Replace(args[1], "!tdm "+usr, "", -1))

				twu.SendDM(e, usr, q)
				break
			}
		case "!tdel":
			{
				twu.DeleteTweet(e, cmd[1])
				break
			}
		case "!tloc":
			{
				q := strings.TrimSpace(strings.Replace(args[1], "!tloc ", "", -1))

				if q == "!tloc" {
					twu.SetLocation(e, "")
				} else {
					twu.SetLocation(e, q)
				}

				break
			}
		case "!tac":
			{
				if len(cmd) > 1 {
					twu.LoadToken(e, cmd[1])
				} else {
					twu.ListTokens(e)
				}
				break
			}
		}
	} else if strings.Index(msg, "youtube.com/") >= 0 {
		utl := new(HttpUtils)
		utl.ParseYoutubeLink(e)
	} else {
		utl := new(HttpUtils)
		utl.GetHttpTitle(e)
		
		for _, v := range rc {
			if v.GetChanName() == args[0] {
				v.SendPrivmsg(e.Message())
			}
		}
	}
}

func main() {
	irc_ready := make(chan int)
	of, ofe := ioutil.ReadFile("options.conf")
	if ofe == nil {
		je := json.Unmarshal(of, &opt)
		if je != nil {
			fmt.Printf("Failed to parse options file: %s\n", je)
		}
	} else {
		fmt.Printf("Options file failed to load: %s\n", ofe)

		//Set default values
		opt = Options{
			Server:          "irc.harkin.me:6667",
			Nick:            "BOT-N",
			useTLS:          true,
			DefaultChans:    []string{"#lobby"},
			OperDetails:     []string{ /* USERNAME, PASSWORD*/ },
			TwitterTokenDir: "./",
		}

		jout, _ := json.Marshal(opt)
		ioutil.WriteFile("options.conf", jout, 0644)
	}

	i := irc.IRC(opt.Nick, opt.Nick)
	i.Debug = true
	i.UseTLS = opt.useTLS
	i.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := i.Connect(opt.Server)
	if err == nil {
		i.AddCallback("001", func(e *irc.Event) {
			//Oper command for SA
			if opt.Oper && len(opt.OperDetails) == 2 {
				i.SendRawf("OPER %s %s", opt.OperDetails[0], opt.OperDetails[1])
			}

			for _, c := range opt.DefaultChans {
				i.Join(c)
				if opt.Oper {
					i.SendRawf("SAMODE %s +o %s", c, opt.Nick)
				}
			}

			irc_ready <- 1
		})
		i.AddCallback("PRIVMSG", OnPrivMsg)
		i.AddCallback("JOIN", OnJoin)
	} else {
		fmt.Printf("Can't connect to: %s\n", opt.Server)
	}

	go func() {
		<-irc_ready
		i.Join("#twitter")
		i.Join("#twitterspam")

		files, err := ioutil.ReadDir(opt.TwitterTokenDir)
		if err == nil {
			var tk = TwitterAuthToken{}

			if len(files) > 0 {
				for _, file := range files {
					of, ofe := ioutil.ReadFile(opt.TwitterTokenDir + "/" + file.Name())
					if ofe == nil {
						je := json.Unmarshal(of, &tk)
						if je == nil {
							fmt.Printf("Starting user stream %s", tk.ScreenName)
							go func() {
								twu.ListenToUserStream(i, tk)
							}()
						} else {
							fmt.Printf("Failed to parse token file %s (%s)", file.Name(), je)
						}
					} else {
						fmt.Printf("Couldn't open token file (%s)", ofe)
					}
				}
			} else {
				fmt.Printf("No twitter tokens found in %s", opt.TwitterTokenDir)
			}
		} else {
			fmt.Printf("Couldn't open token dir (%s)", err)
		}
	}()

	i.Loop()
}
