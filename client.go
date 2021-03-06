package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	server := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
	}
	for {
		before := time.Now()
		//res := Get(server[0])
		//res := Read(server[0], time.Second)
		//res := MultiRead(server, time.Second)
		after := time.Now()
		fmt.Println("Response:", *res)
		fmt.Println("Time:", after.Sub(before))
		fmt.Println()
		time.Sleep(500 * time.Millisecond)
	}
}

type Response struct {
	Body       string
	StatusCode int
}

// Get makes an HTTP Get request and returns an abbreviated response.
// Status code 200 means that the request was successful.
// The function returns &Response{"", 0} if the request fails
// and it blocks forever if the server doesn't respond.
func Get(url string) *Response {
	res, err := http.Get(url)
	if err != nil {
		return &Response{}
	}
	// res.Body != nil when err == nil
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
	}
	return &Response{string(body), res.StatusCode}
}

// FIXME
// I've found two insidious bugs in this function; both of them are unlikely
// to show up in testing. Please fix them right away â€“ and don't forget to
// write a doc comment this time.
// 1. DATA RACE.... Innan kunde res försöka bindas i go-rutinen samtidigt eller ändras till något som den inte ska vara
// senare i select. Löser det med att använda en kanal. Buffrar även kanalen.
// 2. Om båda cases i selecten är sanna (vilket är väldigt osannolikt, men ändå)
// så kommer select kunna välja att köra båda. Vilken som körs väljs 'pseudo-randomly'
// Löser detta genom att res endast kan ge en time-out om res faktiskt är nil. Annars kan det hända att
// res har ett värde men att det tagit exakt så lång tid som det ska ta för att ge en timeout, och då kan skicka ut en
// felaktig time-out

func Read(url string, timeout time.Duration) (res *Response) {
	ch := make(chan *Response, 1)
	go func() {
		ch <- Get(url)
	}()
	select {
	case res = <-ch:
	case <-time.After(timeout):
		if res == nil {
			res = &Response{"Gateway timeout\n", 504}
		} else {
			res = <-ch
		}
	}
	return
}

// MultiRead makes an HTTP Get request to each url and returns
// the response of the first server to answer with status code 200.
// If none of the servers answer before timeout, the response is
// 503 â€“ Service unavailable.
func MultiRead(urls []string, timeout time.Duration) (res *Response) {
	ch := make(chan *Response, (len(urls)))
	for _, x := range urls { // Kör en go-rutin för samtliga strängar i arrayen urls.
		go func(z string) {
			ch <- Get(z)
		}(x)
	}
	select {
	case res = <-ch:
	case <-time.After(timeout):
		if res == nil {
			res = &Response{"Service unavailable", 503} // Service unavailable!
		} else {
			res = <-ch
		}
	}
	return
}
