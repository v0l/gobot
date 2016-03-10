package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/thoj/go-ircevent"
)

type TwitterAuthToken struct {
	OauthToken       string `json:"oauth_token"`
	OauthTokenSecret string `json:"oauth_token_secret"`
	UserID           string `json:"user_id"`
	ScreenName       string `json:"screen_name"`
	XAuthExpires     string `json:"x_auth_expires"`
}

type TweetLocation struct {
	err    error
	result anaconda.GeoSearchResult
	lat    string
	long   string
}

type TwitterUtil struct {
	LocationMode int
	Location     TweetLocation
}

func (*TwitterUtil) GetHelp(e *irc.Event) {
	e.Connection.Privmsgf(e.Arguments[0], "!t <msg> \t- Sends a tweet!")
	e.Connection.Privmsgf(e.Arguments[0], "!tr <id> <msg> \t- Sends a reply to a tweet")
	e.Connection.Privmsgf(e.Arguments[0], "!tf <handle> \t- Follow somebody on twatter")
	e.Connection.Privmsgf(e.Arguments[0], "!tdm <handle> <msg>\t- Send a DM to somebody")
	e.Connection.Privmsgf(e.Arguments[0], "!tloc <location/auto>\t- Sets tweet location")
	e.Connection.Privmsgf(e.Arguments[0], "!tdel <id>\t- Deletes a tweet")
	e.Connection.Privmsgf(e.Arguments[0], "!tac <filename>\t- Sets twitter account")
}

func (*TwitterUtil) ListTokens(e *irc.Event) {
	files, err := ioutil.ReadDir(opt.TwitterTokenDir)
	if err == nil {
		var tk = TwitterAuthToken{}

		if len(files) > 0 {
			e.Connection.Privmsgf(e.Arguments[0], "Twitter account tokens:")
			for _, file := range files {
				of, ofe := ioutil.ReadFile(opt.TwitterTokenDir + "/" + file.Name())
				if ofe == nil {
					je := json.Unmarshal(of, &tk)
					if je == nil {
						if tk.OauthToken == opt.TwitterAuthKey {
							e.Connection.Privmsgf(e.Arguments[0], " - %s (%s)(Active)", tk.ScreenName, file.Name())
						} else {
							e.Connection.Privmsgf(e.Arguments[0], " - %s (%s)", tk.ScreenName, file.Name())
						}
					} else {
						e.Connection.Privmsgf(e.Arguments[0], "Failed to parse token file %s (%s)", file.Name(), je)
					}
				} else {
					e.Connection.Privmsgf(e.Arguments[0], "[%s] Couldn't open token file (%s)", e.Nick, ofe)
				}
			}
		} else {
			e.Connection.Privmsgf(e.Arguments[0], "[%s] No twitter tokens found in %s", e.Nick, opt.TwitterTokenDir)
		}
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Couldn't open token dir (%s)", e.Nick, err)
	}
}

func (*TwitterUtil) LoadToken(e *irc.Event, name string) {
	var tk = TwitterAuthToken{}
	of, ofe := ioutil.ReadFile(opt.TwitterTokenDir + "/" + name)
	if ofe == nil {
		je := json.Unmarshal(of, &tk)
		if je == nil {
			opt.TwitterAuthKey = tk.OauthToken
			opt.TwitterAuthSecret = tk.OauthTokenSecret
			opt.TwitterHandle = tk.ScreenName

			e.Connection.Privmsgf(e.Arguments[0], "[%s] Twitter account set to: %s (%s)", e.Nick, opt.TwitterHandle, tk.OauthToken)
		} else {
			e.Connection.Privmsgf(e.Arguments[0], "[%s] Error parsing json file: %s", e.Nick, name)
		}
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Error reading json file: %s", e.Nick, name)
	}
}

func (t *TwitterUtil) GetNewTweetMedia(txt string) (anaconda.Media, string) {
	rx, _ := regexp.Compile("(http|https):\\/\\/([\\w.\\-\\/\\%#]+)")
	rxm := rx.FindAllStringSubmatch(txt, -1)
	if len(rxm) > 0 {
		for i := 0; i < len(rxm); i++ {
			rxm_i := rxm[i]
			ml := rxm_i[0]

			ht := new(HttpUtils)
			mt := ht.GetContentType(ml)

			if strings.Index(mt, "image") >= 0 {
				anaconda.SetConsumerKey(opt.TwitterAppKey)
				anaconda.SetConsumerSecret(opt.TwitterAppSecret)
				api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

				bimg := ht.GetRemoteImageBase64(ml)
				med, mer := api.UploadMedia(bimg)
				api.Close()
				if mer == nil {
					return med, strings.Replace(txt, ml, "", -1)
				} else {
					fmt.Println(mer)
				}
			}
		}
	}
	return anaconda.Media{}, txt
}

func (t *TwitterUtil) SetLocation(e *irc.Event, txt string) {
	if txt == "" {
		t.Location = TweetLocation{}
		t.LocationMode = 0

		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet location set to: random", e.Nick)
	} else {
		t.LocationMode = 1
		anaconda.SetConsumerKey(opt.TwitterAppKey)
		anaconda.SetConsumerSecret(opt.TwitterAppSecret)
		api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

		vals := url.Values{}
		vals.Add("query", url.QueryEscape(txt))

		pl, ple := api.GeoSearch(vals)
		if ple == nil && len(pl.Result.Places) > 0 && len(pl.Result.Places[0].Centroid) == 2 {
			t.Location.result = pl
			t.Location.err = nil

			t.Location.lat = fmt.Sprintf("%.4f", pl.Result.Places[0].Centroid[1])
			t.Location.long = fmt.Sprintf("%.4f", pl.Result.Places[0].Centroid[0])

			e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet location set to: %s (https://www.google.com/maps?q=%s,%s)", e.Nick, pl.Result.Places[0].FullName, t.Location.lat, t.Location.long)
		} else {
			e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet location error: %s", e.Nick, ple)
		}
		api.Close()
	}
}

func (t *TwitterUtil) GetLoc(api *anaconda.TwitterApi) *TweetLocation {
	if t.LocationMode == 0 {
		rand.Seed(int64(time.Now().Unix()))

		rl := new(TweetLocation)
		rl.lat = fmt.Sprintf("%.4f", (float32(180)*rand.Float32())-float32(90))
		rl.long = fmt.Sprintf("%.4f", (float32(360)*rand.Float32())-float32(180))

		vals := url.Values{}
		vals.Add("lat", rl.lat)
		vals.Add("long", rl.long)

		pl, ple := api.GeoSearch(vals)

		rl.err = ple
		rl.result = pl

		return rl
	} else {
		return &t.Location
	}
}

func (t *TwitterUtil) SendTweet(e *irc.Event, q string) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)
	_, ve := api.VerifyCredentials()
	if ve != nil {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Credentials could not be validated (%s)", e.Nick, ve)
		return
	}

	vals := url.Values{}

	pl := t.GetLoc(api)
	if pl.err == nil {
		vals.Add("lat", pl.lat)
		vals.Add("long", pl.long)
		vals.Add("place_id", pl.result.Result.Places[0].ID)
	}

	tm, twe := t.GetNewTweetMedia(q)
	if tm.MediaID != 0 {
		vals.Add("media_ids", tm.MediaIDString)
	}
	tw, ter := api.PostTweet(twe, vals)
	if ter == nil {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet sent [%s][ https://twitter.com/%s/status/%s ]", e.Nick, tw.IdStr, tw.User.ScreenName, tw.IdStr)
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet send failed: %s", e.Nick, ter)
	}

	api.Close()
}

func (t *TwitterUtil) SendTweetResponse(e *irc.Event, q string, tid string) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

	vals := url.Values{}

	pl := t.GetLoc(api)
	if pl.err == nil {
		vals.Add("lat", pl.lat)
		vals.Add("long", pl.long)
		vals.Add("place_id", pl.result.Result.Places[0].ID)
	}

	vals.Add("in_reply_to_status_id", tid)

	tm, twe := t.GetNewTweetMedia(q)
	if tm.MediaID != 0 {
		vals.Add("media_ids", tm.MediaIDString)
	}
	tw, ter := api.PostTweet(twe, vals)
	if ter == nil {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet sent [ https://twitter.com/%s/status/%s ]", e.Nick, tw.User.ScreenName, tw.IdStr)
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet send failed: %s", e.Nick, ter)
	}

	api.Close()
}

func (*TwitterUtil) Follow(e *irc.Event, usr string) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

	tw, ter := api.FollowUser(usr)
	if ter == nil {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Now following @%s", e.Nick, tw.ScreenName)
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet send failed: %s", e.Nick, ter)
	}

	api.Close()
}

func (*TwitterUtil) SendDM(e *irc.Event, usr, q string) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

	dm, ter := api.PostDMToScreenName(q, usr)
	if ter == nil {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] DM sent to @%s", e.Nick, dm.RecipientScreenName)
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] DM send failed: %s", e.Nick, ter)
	}
	api.Close()
}

func (*TwitterUtil) DeleteTweet(e *irc.Event, tid string) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

	id, ier := strconv.ParseInt(tid, 10, 64)
	if ier == nil {
		_, er := api.DeleteTweet(id, false)
		if er == nil {
			e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet deleted", e.Nick)
		} else {
			e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet delete failed: %s", e.Nick, er)
		}
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet id is invalid int64", e.Nick)
	}

	api.Close()
}

func (*TwitterUtil) ListenToUserStream(i *irc.Connection, auth, secret string) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(auth, secret)

	stream := api.UserStream(url.Values{})
	if stream != nil {
		i.Join("#twitter")
		i.Join("#twitterspam")

		for {
			tw := <-stream.C
			switch st := tw.(type) {
			case anaconda.Tweet:
				{
					//Check if the tweet is at us, otherwise spam #twitterspam
					hasMention := false
					for x := 0; x < len(st.Entities.User_mentions); x++ {
						ent := st.Entities.User_mentions[x]
						if ent.Screen_name == opt.TwitterHandle {
							hasMention = true
						}
					}

					tweet_txt := strings.Replace(st.Text, "\n", "", -1)

					if hasMention {
						i.Privmsgf("#twitter", "Tweet from @%s [%s]: %s", st.User.ScreenName, st.IdStr, tweet_txt)
					} else {
						nc.Chat(tweet_txt)
						i.Privmsgf("#twitterspam", "@%s [%s]: %s", st.User.ScreenName, st.IdStr, tweet_txt)
					}
					break
				}
			case anaconda.DirectMessage:
				{
					dm_txt := strings.Replace(st.Text, "\n", "", -1)
					if st.SenderScreenName != opt.TwitterHandle {
						i.Privmsgf("#twitter", "DM from @%s [%s]: %s", st.SenderScreenName, st.IdStr, dm_txt)
					}
					break
				}
			case anaconda.EventTweet:
				{
					tweet_txt := strings.Replace(st.TargetObject.Text, "\n", "", -1)
					if st.Event.Event == "favorite" && st.Source.ScreenName != opt.TwitterHandle {
						i.Privmsgf("#twitter", "Tweet favorited by @%s [%s]: %s", st.Source.ScreenName, tweet_txt)
					}
					break
				}
			case anaconda.Event:
				{
					if st.Event == "follow" && st.Source.ScreenName != opt.TwitterHandle {
						i.Privmsgf("#twitter", "@%s is now following you", st.Source.ScreenName)
					}
					break
				}
			}
		}
	}

	api.Close()
}
