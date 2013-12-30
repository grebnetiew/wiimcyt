// Serves an HTTP server which accepts requests of the form
// document?q=things
// and then searches for things using the youtube api.
// The results are converted to pls format and returned.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Make the http request to youtube's api
	resp, err := http.Get("https://gdata.youtube.com/feeds/api/videos?q=" + query)
	defer resp.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("http.Get(youtube): ", err)
		return
	}
	// Parse urls from the blob of xml
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Reading response: ", err)
		return
	}
	re := regexp.MustCompile(`<entry>.*?<title.*?>(.*?)</title>.*?(https?://www\.youtube\.com/watch\?v=.{10,20}&amp;feature=youtube_gdata)`)
	entries := re.FindAllSubmatch(respBytes, 99)

	// Send the result back to the client
	w.Header()["Content-Type"] = []string{"audio/x-scpls"}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[playlist]\n"))
	w.Write([]byte(fmt.Sprintf("NumberOfEntries=%d\n", len(entries))))
	for index, entry := range entries {
		tstr := string(entry[1])
		ustr := string(entry[2])
		ustr = strings.Replace(string(ustr), "&amp;", "&", -1)
		ustr = strings.Replace(string(ustr), "https:", "http:", 1)
		w.Write([]byte(fmt.Sprintf("File%d=%s\n", index+1, ustr)))
		w.Write([]byte(fmt.Sprintf("Title%d=%s\n", index+1, tstr)))
	}
	log.Println("Successful response to query '", query, "'")
}

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8089", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
