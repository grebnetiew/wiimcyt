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
	Title Title
	Link  []Link
	Media Media `json:"media$group"`
}
type Media struct {
	Thumb []Thumb `json:"media$thumbnail"`
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

// Settings
const (
	// Which user's feed to download on empty query
	ytUser = "kire456"
	// Which port to serve on
	httpServerAddr = ":8089"
)

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	var resp *http.Response
	var err error
	if query == " " {
		// Special feature: on empty query, return someone's subscriptions
		log.Println("Responding to request for new videos")
		resp, err = http.Get("https://gdata.youtube.com/feeds/api/users/" +
			ytUser + "/newsubscriptionvideos")
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
		title := entry.Title.Text
		// WiiMC doesn't understand https
		video := strings.Replace(SelectAlternateLink(entry.Link).Url,
			"https:", "http:", 1)
		//thumb := strings.Replace(SelectBigThumbnail(entry.Media.Thumb).Url,
		//	"https:", "http:", 1)
		w.Write([]byte(fmt.Sprintf("File%d=%s\n", index+1, video)))
		w.Write([]byte(fmt.Sprintf("Title%d=%s\n", index+1, title)))
		//w.Write([]byte(fmt.Sprintf("Thumbnail%d=%s\n", index+1, thumb)))
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
