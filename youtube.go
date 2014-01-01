// Serves an HTTP server which accepts requests of the form
// document?q=things
// and then searches for things using the youtube api.
// The results are converted to pls format and returned.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Madness required for correctly parsing youtube's api response
type YTFeed struct {
	Feed Feed
}
type Feed struct {
	Entries []Entry `json:"entry"`
}
type Entry struct {
	Title  Title
	Author []Author
	Link   []Link
	Media  Media `json:"media$group"`
}
type Author struct {
	Name Title
}
type Media struct {
	Thumb    []Thumb  `json:"media$thumbnail"`
	Duration Duration `json:"yt$duration"`
}
type Duration struct {
	Seconds string
}
type Title struct {
	Text string `json:"$t"`
}
type Link struct {
	Rel string
	Url string `json:"href"`
}
type Thumb struct {
	Url    string
	Width  int
	Height int
}

type Video struct {
	Author, Title, Link, Display string
	Duration            int
	Thumb               string
}

// Settings
const (
	// Which user's feed to download on empty query
	ytUser = "kire456"
	// Which port to serve on
	httpServerAddr = ":8089"
	// Set to true after installing a custom unicode font on the wii
	supportUnicode = false
)

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	var resp *http.Response
	var err error
	if query == " " {
		// Special feature: on empty query, return someone's subscriptions
		log.Println("Responding to request for new videos")
		resp, err = http.Get("https://gdata.youtube.com/feeds/api/users/" +
			ytUser + "/newsubscriptionvideos?alt=json")
	} else {
		log.Println("Responding to query '" + query + "'")
		// Make the http request to youtube's api
		resp, err = http.Get("https://gdata.youtube.com/feeds/api/videos" +
			"?alt=json&q=" + url.QueryEscape(query))
	}
	defer resp.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("http.Get(youtube): ", err)
		return
	}
	// Parse entries from the blob of xml
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Reading response: ", err)
		return
	}
	var yt YTFeed
	err = json.Unmarshal(respBytes, &yt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("json: ", err)
		return
	}

	// Send the result back to the client
	w.Header()["Content-Type"] = []string{"audio/x-scpls"}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[playlist]\n"))
	w.Write([]byte(fmt.Sprintf("NumberOfEntries=%d\n", len(yt.Feed.Entries))))
	for index, entry := range yt.Feed.Entries {
		v := entry.Parse()
		w.Write([]byte(fmt.Sprintf("File%d=%s\n", index+1, v.Link)))
		w.Write([]byte(fmt.Sprintf("Title%d=%s\n", index+1, v.Display)))
		w.Write([]byte(fmt.Sprintf("Length%d=%d\n", index+1, v.Duration)))
	}
}

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(httpServerAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// There are many links in the feed and most are not the video.
// The video has the rel attribute set to alternate.
func SelectAlternateLink(links []Link) Link {
	if len(links) == 0 {
		return Link{"", ""}
	}
	for _, link := range links {
		if link.Rel == "alternate" {
			return link
		}
	}
	return links[0]
}

// Thumbnails come in two sizes, small (90px) and large (360).
// We'd like the big one for display on the TV.
func SelectBigThumbnail(thumbs []Thumb) Thumb {
	if len(thumbs) == 0 {
		return Thumb{"", 0, 0}
	}
	for _, thumb := range thumbs {
		if thumb.Width > 200 {
			return thumb
		}
	}
	return thumbs[0]
}

func (e *Entry) Parse() *Video {
	var duration int
	fmt.Sscanf(e.Media.Duration.Seconds, "%d", &duration)
	// Displayed title doesn't contain non-ascii, since WiiMC doesn't
	// display that correctly
	display := []rune("[" + e.Author[0].Name.Text + "] " + e.Title.Text)
	for i := range display {
		if supportUnicode || display[i] > 255 {
			display[i] = 164 // currency mark, slightly block-shaped 
		}
	}
	return &Video{
		Author: e.Author[0].Name.Text,
		Title:  e.Title.Text,
		Display: string(display),
		// WiiMC doesn't understand https
		Link: strings.Replace(SelectAlternateLink(e.Link).Url,
			"https:", "http:", 1),
		Thumb: strings.Replace(SelectBigThumbnail(e.Media.Thumb).Url,
			"https:", "http:", 1),
		Duration: duration,
	}
}
