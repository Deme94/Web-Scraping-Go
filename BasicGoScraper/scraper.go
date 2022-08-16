package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Initialize the client and set the cookies
	client := NewHttpClient()
	// Url
	u := url.URL{
		Host: "https://stadia.google.com/games",
	}
	// Cookies
	c := http.Cookie{
		Name:  "SessionId_example",
		Value: "12345678_example",
		//MaxAge: default number,
	}
	var cookies []*http.Cookie
	cookies = append(cookies, &c)

	client.Jar.SetCookies(&u, cookies)

	// Request the HTML page.
	res, err := client.Get(u.Host)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var games []Game
	// Find all games (excludes free and pro games)
	doc.Find(".d5UsQb div > picture > img").Each(func(i int, allGames *goquery.Selection) {
		gameTitle, _ := allGames.Attr("alt")
		gameImageURL, _ := allGames.Attr("src")
		games = append(games, Game{gameImageURL, gameTitle})
	})
	gamesJson, err := json.Marshal(games)
	if err != nil {
		panic(err)
	}
	gamesBody := bytes.NewBuffer(gamesJson)
	fmt.Print(gamesBody)
}
