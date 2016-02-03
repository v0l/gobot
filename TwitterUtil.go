package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/thoj/go-ircevent"
)

type TwitterUtil struct {
}

func (*TwitterUtil) GetHelp(e *irc.Event) {
	e.Connection.Privmsgf(e.Arguments[0], "!t <msg> \t- Sends a tweet!")
	e.Connection.Privmsgf(e.Arguments[0], "!tr <id> <msg> \t- Sends a reply to a tweet")
	e.Connection.Privmsgf(e.Arguments[0], "!tf <handle> \t- Follow somebody on twatter")
	e.Connection.Privmsgf(e.Arguments[0], "!tdm <handle> <msg>\t- Send a DM to somebody")
}

func (*TwitterUtil) SendTweet(e *irc.Event, q string) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

	vals := url.Values{}
	rand.Seed(int64(time.Now().Unix()))
	vals.Add("lat", fmt.Sprintf("%.4f", (float32(180)*rand.Float32())-float32(90)))
	vals.Add("long", fmt.Sprintf("%.4f", (float32(360)*rand.Float32())-float32(180)))

	pl, ple := api.GeoSearch(vals)
	if ple == nil {
		vals.Add("place_id", pl.Result.Places[0].ID)
	}

	tw, ter := api.PostTweet(q, vals)
	if ter == nil {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet sent [ https://twitter.com/%s/status/%s ]", e.Nick, tw.User.ScreenName, tw.IdStr)
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] Tweet send failed: %s", e.Nick, ter)
	}

	api.Close()
}

func (*TwitterUtil) SendTweetResponse(e *irc.Event, q string, tid string) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

	vals := url.Values{}
	rand.Seed(int64(time.Now().Unix()))
	vals.Add("lat", fmt.Sprintf("%.4f", (float32(180)*rand.Float32())-float32(90)))
	vals.Add("long", fmt.Sprintf("%.4f", (float32(360)*rand.Float32())-float32(180)))

	pl, ple := api.GeoSearch(vals)
	if ple == nil {
		vals.Add("place_id", pl.Result.Places[0].ID)
	}
	vals.Add("in_reply_to_status_id", tid)

	tw, ter := api.PostTweet(q, vals)
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

	dm, ter := api.PostDMToScreenName(usr, q)
	if ter == nil {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] DM sent to @%s", e.Nick, dm.RecipientScreenName)
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "[%s] DM send failed: %s", e.Nick, ter)
	}
	api.Close()
}

func (*TwitterUtil) ListenToUserStream(i *irc.Connection) {
	anaconda.SetConsumerKey(opt.TwitterAppKey)
	anaconda.SetConsumerSecret(opt.TwitterAppSecret)
	api := anaconda.NewTwitterApi(opt.TwitterAuthKey, opt.TwitterAuthSecret)

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
						i.Privmsgf("#twitterspam", "@%s [%s]: %s", st.User.ScreenName, st.IdStr, tweet_txt)
					}
					break
				}
			case anaconda.DirectMessage:
				{
					dm_txt := strings.Replace(st.Text, "\n", "", -1)
					i.Privmsgf("#twitter", "DM from @%s [%s]: %s", st.SenderScreenName, st.IdStr, dm_txt)
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
