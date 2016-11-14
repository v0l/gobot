package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/domainr/whois"
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
	TwitterAppKey      string
	TwitterAppSecret   string
	TwitterTokenDir    string
	TwitterStreamToken string

	//Temp twitter vars
	TwitterAuthKey    string
	TwitterAuthSecret string
	TwitterHandle     string

	//Tropo
	TropoToken    string
	TropoTokenMsg string
	CallDir       string

	//Email
	MailServer   string
	MailUser     string
	MailPassword string

	//Google
	GoogleNID string
	GoogleDV string
	GoogleConsent string
	UserAgent string
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
				e.Connection.Privmsgf(args[0], "!callg <no> \t- Calls a phone number list and confrences them eg. !callg number1,number2,number3")
				e.Connection.Privmsgf(args[0], "!sms <no> <msg> \t- Send an SMS to the supplied phone number")
				e.Connection.Privmsgf(args[0], "!ud (!w) <word>\t- Looks words up on urban dictionary")
				e.Connection.Privmsgf(args[0], "!s <thing>\t- Does a quick Google")
				e.Connection.Privmsgf(args[0], "!js <code> \t- Runs some JS code")
				e.Connection.Privmsgf(args[0], "!remote <server> <nick> <chan> <ssl>\t- Connectes to another IRC server and pipes chat to #lobby")
				e.Connection.Privmsgf(args[0], "!rclose <server#>\t- Close connection to remote")
				e.Connection.Privmsgf(args[0], "!rjoin <server#> <chan>\t- Joins another chan on a remote connection")
				e.Connection.Privmsgf(args[0], "!ip <dns> \t- Gets ip addresses for domain")
				e.Connection.Privmsgf(args[0], "!thelp \t- Gets twitter command list")
				break
			}
		case "!whois":
			{
				if len(cmd) > 1 {
					request, err := whois.NewRequest(cmd[1])
					if err == nil {
						response, err2 := whois.DefaultClient.Fetch(request)
						if err2 == nil {
							rs := strings.Split(response.String(), "\n")
							for _, v := range rs {
								if strings.Index(v, "%") == -1 {
									if (strings.Index(v, ":") != -1 && strings.Index(v, ":") <= len(v)-4) || strings.Index(v, ":") == -1 {
										e.Connection.Privmsgf(args[0], "%v", v)
									}
								}
							}
						} else {
							e.Connection.Privmsgf(args[0], "Error: %v", err2)
						}
					} else {
						e.Connection.Privmsgf(args[0], "Error: %v", err)
					}
				}
				break
			}
		case "!end":{
				end := time.Date(2066, time.June, 6, 12, 34, 56, 0, time.UTC)
				dl := end.Sub(time.Now())
				
				dld := dl.Hours() / 24
				e.Connection.Privmsgf(args[0], "[%v] %.2f days until you no longer exist.", e.Nick, dld)
				break
			}
		case "!ip":
			{
				if len(cmd) > 1 {
					ips, er := net.LookupIP(cmd[1])
					if er == nil {
						for _, ip := range ips {
							e.Connection.Privmsgf(args[0], "[%v] %v %v", e.Nick, ip.String(), cmd[1])
						}
					}
				}
				break
			}
		case "!8":
			{
				if len(cmd) > 1 {
					ans := []string{"It is certain", "It is decidedly so", "Without a doubt", "Yes, definitely", "You may rely on it", "As I see it, yes", "Most likely", "Outlook good", "Yes", "Signs point to yes", "Reply hazy try again", "Ask again later", "Better not tell you now", "Cannot predict now", "Concentrate and ask again", "Don't count on it", "My reply is no", "My sources say no", "Outlook not so good", "Very doubtful"}
					rand.Seed(time.Now().UnixNano())
					e.Connection.Privmsgf(args[0], "%v: %v", e.Nick, ans[rand.Intn(len(ans))])
				}
				break
			}
		case "!remote":
			{
				if len(cmd) > 4 {
					nc := RemoteIrc{}
					nc.Main = e.Connection
					ncs := false
					if strings.ToLower(cmd[4]) == "true" {
						ncs = true
					}

					if nc.Run(cmd[1], cmd[2], cmd[3], ncs) {
						rc = append(rc, nc)

						js, jer := json.Marshal(rc)
						if jer == nil {
							ioutil.WriteFile("remote.json", js, 0644)
						}

						e.Connection.Privmsgf(args[0], "[%v] Connected to %s", e.Nick, cmd[1])
					}
				} else {
					e.Connection.Privmsgf(args[0], "Remote connections:")
					for k, v := range rc {
						e.Connection.Privmsgf(args[0], "%v: %s %s (%s)", k, v.Server, v.Channels, v.Nick)
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

							js, jer := json.Marshal(rc)
							if jer == nil {
								ioutil.WriteFile("remote.json", js, 0644)
							}

							defer tc.Stop()
						}
					}
				}
				break
			}
		case "!rjoin":
			{
				if len(cmd) > 2 {
					srv, ser := strconv.Atoi(cmd[1])
					if ser == nil {
						if srv < len(rc) {
							tc := rc[srv]
							tc.JoinChan(cmd[2])

							js, jer := json.Marshal(rc)
							if jer == nil {
								ioutil.WriteFile("remote.json", js, 0644)
							}
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
				q := url.QueryEscape(strings.TrimSpace(strings.Replace(args[1], "!s ", "", -1)))
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
		case "!callg":
			{
				pn := cmd[1]

				tr := new(Tropo)
				tr.CallGroup(e, pn)
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
					twu.LoadTokenCmd(e, cmd[1])
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
			ch := args[0][3:]
			if v.HasChan(ch) {
				v.SendPrivmsg(e.Message(), ch)
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

		tk := twu.LoadToken(opt.TwitterStreamToken)
		twu.ListenToUserStream(i, tk)

		//load remote connections
		rf, fer := ioutil.ReadFile("remote.json")
		if fer == nil {
			jer := json.Unmarshal(rf, &rc)
			if jer != nil {
				i.Privmsgf("#lobby", "Failed to load remote connection list, %s", jer.Error())
			}
		}

		for _, r := range rc {
			r.Main = i
			r.Start()
		}
	}()

	i.Loop()
}
