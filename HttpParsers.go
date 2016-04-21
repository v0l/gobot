package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/thoj/go-ircevent"
	"golang.org/x/net/html"
)

type GoogleSearch struct {
	ResponseData struct {
		Results []struct {
			GsearchResultClass string `json:"GsearchResultClass"`
			UnescapedURL       string `json:"unescapedUrl"`
			URL                string `json:"url"`
			VisibleURL         string `json:"visibleUrl"`
			CacheURL           string `json:"cacheUrl"`
			Title              string `json:"title"`
			TitleNoFormatting  string `json:"titleNoFormatting"`
			Content            string `json:"content"`
		} `json:"results"`
		Cursor struct {
			ResultCount string `json:"resultCount"`
			Pages       []struct {
				Start string `json:"start"`
				Label int    `json:"label"`
			} `json:"pages"`
			EstimatedResultCount string `json:"estimatedResultCount"`
			CurrentPageIndex     int    `json:"currentPageIndex"`
			MoreResultsURL       string `json:"moreResultsUrl"`
			SearchResultTime     string `json:"searchResultTime"`
		} `json:"cursor"`
	} `json:"responseData"`
	ResponseDetails interface{} `json:"responseDetails"`
	ResponseStatus  int         `json:"responseStatus"`
}

type YtData struct {
	Kind     string `json:"kind"`
	Etag     string `json:"etag"`
	Pageinfo struct {
		Totalresults   int `json:"totalResults"`
		Resultsperpage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []struct {
		Kind    string `json:"kind"`
		Etag    string `json:"etag"`
		ID      string `json:"id"`
		Snippet struct {
			Publishedat string `json:"publishedAt"`
			Channelid   string `json:"channelId"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"default"`
				Medium struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"medium"`
				High struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"high"`
				Standard struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"standard"`
				Maxres struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"maxres"`
			} `json:"thumbnails"`
			Channeltitle         string `json:"channelTitle"`
			Categoryid           string `json:"categoryId"`
			Livebroadcastcontent string `json:"liveBroadcastContent"`
			Localized            struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"localized"`
		} `json:"snippet"`
	} `json:"items"`
}

const (
	USERAGENT string = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.97 Safari/537.36"
)

type HttpUtils struct {
}

func (*HttpUtils) ParseYoutubeLink(e *irc.Event) {
	msg := e.Message()

	ytr, _ := regexp.Compile("\\?v=(\\w+)")
	ytm := ytr.FindAllStringSubmatch(msg, -1)
	if len(ytm) > 0 {
		for _, y := range ytm {
			//Load video data
			apiKey := opt.YoutubeApiKey
			doc, de := http.Get(fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet&id=%s&key=%s", y[1], apiKey))
			if de == nil {
				defer doc.Body.Close()

				var gd YtData
				body, _ := ioutil.ReadAll(doc.Body)
				jer := json.Unmarshal(body, &gd)
				if jer == nil {
					if len(gd.Items) > 0 {
						e.Connection.Privmsgf(e.Arguments[0], "\u25B2 %s \u25B2", gd.Items[0].Snippet.Title)
					} else {
						e.Connection.Privmsgf(e.Arguments[0], "Nothing found for %s", y[1])
					}
				}
			}
		}
	} else {
		e.Connection.Privmsgf(e.Arguments[0], "(%s) Malformed youtube link!", e.Nick)
	}
}

func (*HttpUtils) GetLove(e *irc.Event) {
	doc_p, _ := http.NewRequest("GET", "http://thecodinglove.com/random", nil)
	doc, de := http.DefaultTransport.RoundTrip(doc_p)

	if de == nil {
		if doc.StatusCode == 302 {
			e.Connection.Privmsgf(e.Arguments[0], "[%s] %s", e.Nick, doc.Header["Location"][0])
		} else {
			e.Connection.Privmsgf(e.Arguments[0], "[%s] Uhh something went wrong.. no love for you!", e.Nick)
		}
	}
}

func (*HttpUtils) GetExcuse(e *irc.Event) {
	doc, de := http.Get("http://programmingexcuses.com/")
	if de == nil {
		defer doc.Body.Close()

		z, _ := html.Parse(doc.Body)
		done := false
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" && !done {
				exc := n.FirstChild.Data
				e.Connection.Privmsgf(e.Arguments[0], "\"%s\"", exc)
				done = true
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}

		f(z)
	}
}

func (*HttpUtils) GetHttpTitle(e *irc.Event) {
	var title []string
	title_regex, re := regexp.Compile("(http|https):\\/\\/([a-zA-Z0-9.\\/?=&_%\\-#]+)")
	if re == nil {
		title = title_regex.FindStringSubmatch(e.Message())
	}

	//Generic title scraper
	if len(title) == 3 {
		doc, de := http.Get(fmt.Sprintf("%s://%s", title[1], title[2]))
		if de == nil {
			defer doc.Body.Close()
			body, _ := ioutil.ReadAll(doc.Body)

			var tl string
			b := string(body)
			start := strings.Index(b, "<title>")
			end := strings.Index(b, "</title>")
			if start != -1 && end != -1 && end > start {
				tl, _ = url.QueryUnescape(strings.TrimSpace(b[start+7 : end]))
				e.Connection.Privmsgf(e.Arguments[0], "\u25B2 %s \u25B2", tl)
			}
		}
	}
}

func (*HttpUtils) SearchGoogle(e *irc.Event, q string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.google.ie/search?q=%s&gws_rd=ssl", url.QueryEscape(q)), nil)
	req.Header.Set("User-Agent", USERAGENT)
	req.Header.Set("dnt", "1")

	doc, de := client.Do(req)
	if de == nil {
		defer doc.Body.Close()

		sr := GoogleSearch{}
		data, re := ioutil.ReadAll(doc.Body)
		if re == nil {
			jer := json.Unmarshal(data, &sr)
			if jer == nil {
				if sr.ResponseStatus == 200 && len(sr.ResponseData.Results) > 0 {
					res := sr.ResponseData.Results[0]

					e.Connection.Privmsgf(e.Arguments[0], "[%s]: (%s) - %s", e.Nick, res.TitleNoFormatting, res.URL)
				}
			}
		}
	}
}

func (*HttpUtils) SearchUrbanDictionary(e *irc.Event, q string) {
	doc, de := http.Get(fmt.Sprintf("http://www.urbandictionary.com/define.php?term=%s", url.QueryEscape(q)))
	if de == nil {
		defer doc.Body.Close()

		z, _ := html.Parse(doc.Body)
		done := false
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "div" && !done {
				for _, a := range n.Attr {
					if a.Key == "class" && a.Val == "meaning" {
						text := ""
						for cn := n.FirstChild; cn != nil; cn = cn.NextSibling {
							if cn.Data == "a" {
								text += strings.Replace(cn.FirstChild.Data, "\n", "", -1)
							} else if cn.Data == "br" {
								text += " "
							} else {
								text += strings.Replace(cn.Data, "\n", "", -1)
							}
						}

						if len(text) > 140 {
							e.Connection.Privmsgf(e.Arguments[0], "[(%s) - %s]: %s", e.Nick, q, text[0:140])
							for x := 0; x < (len(text)-140)/140; x++ {
								st := 140 + (x * 140)
								if st+140 > len(text) {
									e.Connection.Privmsgf(e.Arguments[0], "%s", text[st:len(text)-st])
								} else {
									e.Connection.Privmsgf(e.Arguments[0], "%s", text[st:st+140])
								}

							}
						} else {
							e.Connection.Privmsgf(e.Arguments[0], "[(%s) - %s]: %s", e.Nick, q, text)
						}

						done = true
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(z)
	}
}

func (*HttpUtils) GetFunFact(e *irc.Event) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://www.google.ie/search?q=im+feeling+curious&gws_rd=ssl", nil)
	req.Header.Set("User-Agent", USERAGENT)
	doc, de := client.Do(req)
	if de == nil {
		defer doc.Body.Close()

		z, _ := html.Parse(doc.Body)
		done := false
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && !done {
				for _, a := range n.Attr {
					if a.Key == "class" && a.Val == "_fXg mod" {
						question := n.FirstChild.NextSibling.FirstChild.FirstChild.Data
						answer := n.FirstChild.NextSibling.FirstChild.NextSibling.FirstChild.NextSibling.FirstChild.Data

						e.Connection.Privmsgf(e.Arguments[0], "%s - %s", e.Nick, question)
						e.Connection.Privmsgf(e.Arguments[0], "%s", answer)

						done = true
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(z)
	}
}

func (*HttpUtils) GetContentType(url string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("User-Agent", USERAGENT)
	doc, de := client.Do(req)
	if de == nil {
		defer doc.Body.Close()
		return doc.Header.Get("Content-Type")
	}

	return ""
}

func (*HttpUtils) GetRemoteImageBase64(url string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", USERAGENT)
	doc, de := client.Do(req)
	if de == nil {
		defer doc.Body.Close()

		img, re := ioutil.ReadAll(doc.Body)
		if re == nil {
			return base64.StdEncoding.EncodeToString(img)
		}
	}

	return ""
}

func (*HttpUtils) GetBodyText(url string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", USERAGENT)
	doc, de := client.Do(req)
	if de == nil {
		defer doc.Body.Close()

		img, re := ioutil.ReadAll(doc.Body)
		if re == nil {
			return string(img)
		}
	}

	return ""
}
