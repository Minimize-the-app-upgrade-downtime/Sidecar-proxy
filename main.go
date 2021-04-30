package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var reconsile = false
var queue []*http.Request

type Handler struct {
}

func stored(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request pushed into queue[%d]", len(queue))
	resp := w

	queue = append(queue, r)
	time.Sleep(15 * time.Second)

	popQueue(resp)
	// This will return http.ErrHandlerTimeout
	log.Println(w.Write([]byte("body")))
}

func popQueue(w http.ResponseWriter) {
	//	var w http.ResponseWriter =  { 0 0}

	for len(queue) > 0 {
		log.Println(queue[0])
		server(w, queue[0])
		queue = queue[1:]
	}
}

func server(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	fmt.Println()
	fmt.Println()
	// fmt.Printf("%S", w)

	link := "http://localhost:3000"
	// parse url
	url, _ := url.Parse(link)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	fmt.Println(proxy)
	// serveHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w, r)

}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if reconsile {
		server(w, r)
	} else {
		stored(w, r)
	}

}

func main() {

	fmt.Println("Server is up!.")
	h := &Handler{} // call the serverHTTP
	// response msg
	http.Handle("/", http.TimeoutHandler(h, 30*time.Second, "500, Internal error"))
	http.ListenAndServe(":50000", nil)

}
