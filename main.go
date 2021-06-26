package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var expectedSleepTime = 60 * time.Second
var isUpdated = true // Nothing update going on.

type handle struct {
	reqQueue  []http.Request
	respQueue []http.ResponseWriter
}

type ESleepTime struct {
	Etime int
}

func (h *handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isUpdated {
		// redirect to application
		link := "http://localhost:3000"
		url, _ := url.Parse(link)
		proxy := httputil.NewSingleHostReverseProxy(url) // create a reverse proxy
		proxy.ServeHTTP(w, r)
	} else {
		// store in the Queue, Update is going on.
		log.Printf("Request pushed into reqQueue[%d]", len(h.reqQueue))
		h.reqQueue = append(h.reqQueue, *r)
		h.respQueue = append(h.respQueue, w)
		time.Sleep(expectedSleepTime)
		w.Write([]byte("500, Internal Server error"))
	}
}

// dequeue the request and send to application
func (h *handle) deQueue(w http.ResponseWriter, r *http.Request) {
	log.Println("Dequeue start...")
	for len(h.reqQueue) > 0 {

		req := h.reqQueue[0]
		resp := h.respQueue[0]
		// accoding to request type divided
		switch req.Method {
		case "GET":
			log.Println("GET request pass to port:3000")
			link := "http://localhost:3000" + req.URL.String()
			r, err := http.Get(link)
			if err != nil {
				log.Println("GET method request err : ", err)
			}
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("response body read error : ", err)
			}
			resp.Write(body)

		case "POST", "PUT", "DELETE":
			log.Println("Request pass to port: 50002")
			link := "http://localhost:50002"
			url, _ := url.Parse(link)
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxy.ServeHTTP(resp, &req)
		default:
			log.Println("Unauthoried request!!. ", req.Method)
		}
		h.reqQueue = h.reqQueue[1:]   // remove request
		h.respQueue = h.respQueue[1:] // remove response
	}
}

func main() {
	log.Println("server is up.")
	h := &handle{}
	http.Handle("/", http.TimeoutHandler(h, (expectedSleepTime+10*time.Second), "500, Internal error"))

	http.HandleFunc("/updateStarted", func(_ http.ResponseWriter, _ *http.Request) {
		isUpdated = false
		log.Println("Ready for the updated Request enQueue", isUpdated)
	})
	http.HandleFunc("/updateFinished", func(resp http.ResponseWriter, req *http.Request) {
		isUpdated = true
		log.Println("Update finished successfully! deQueue is started. ", isUpdated)
		h.deQueue(resp, req)
	})
	http.HandleFunc("/expcedTime", func(resp http.ResponseWriter, req *http.Request) {

		reqBody, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			log.Println("req body read error : ", err)
		}
		log.Println("read req body : ", string(reqBody))
		var t ESleepTime
		err = json.Unmarshal(reqBody, &t)
		if err != nil {
			log.Println("Unmarshal error : ", err)
		}
		expectedSleepTime = time.Duration(t.Etime) * time.Second
		log.Println("Expected Update time : ", expectedSleepTime)
	})
	http.ListenAndServe(":50000", nil)
}
