package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var reconsile = true  // if ture cluster working. false cluster reconsile
var queue []*http.Request
var resqueue []http.ResponseWriter 


type Handler struct {
}


// request store in queue cluster is updating
func storeReqInQueue(w http.ResponseWriter, r *http.Request) {
	
	log.Printf("Request pushed into queue[%d]", len(queue))

	queue = append(queue, r)  // add new req into queue
	resqueue = append(resqueue, w) // add new res into queue
	time.Sleep(50 *time.Second) // 60s wait 
	
	// This will return http.ErrHandlerTimeout
	w.Write([]byte("500, Internal Server error"))
}


// pop request queue until queue is empty
func popReqQueue(res http.ResponseWriter, req *http.Request) {
	

	for len(queue) > 0 {
		//log.Println(queue[0])
		r := queue[0] // req
		w := resqueue[0]
	//	link := "http://localhost:3000"
		
		switch r.Method{
			
			case "GET" :	
				log.Println("GET request pass to localhost:3000 with downtime")
				link := "http://localhost:3000"
				// parse url
				url, _ := url.Parse(link)
				// create the reverse proxy
				proxy := httputil.NewSingleHostReverseProxy(url)
				//fmt.Println(proxy)
				// serveHttp is non blocking and uses a go routine under the hood
				proxy.ServeHTTP(w, r)

			case "POST" :
				log.Println("POST request pass to localhost:50002 with downtime")
				link := "http://localhost:50002/"
				// parse url
				url, _ := url.Parse(link)
				// create the reverse proxy
				proxy := httputil.NewSingleHostReverseProxy(url)
				//fmt.Println(proxy)
				// serveHttp is non blocking and uses a go routine under the hood
				proxy.ServeHTTP(w, r)
			
			case "PUT" :
				log.Println("PUT request pass to localhost:50002 with downtime")
				link := "http://localhost:50002/"
				// parse url
				url, _ := url.Parse(link)
				// create the reverse proxy
				proxy := httputil.NewSingleHostReverseProxy(url)
				//fmt.Println(proxy)
				// serveHttp is non blocking and uses a go routine under the hood
				proxy.ServeHTTP(w, r)
			
			case "DELETE" :
				log.Println("DELETE request pass to localhost:50002 with downtime")
				link := "http://localhost:50002/"
				// parse url
				url, _ := url.Parse(link)
				// create the reverse proxy
				proxy := httputil.NewSingleHostReverseProxy(url)
				//fmt.Println(proxy)
				// serveHttp is non blocking and uses a go routine under the hood
				proxy.ServeHTTP(w, r)
			
			default:
				log.Println("reqest type not found")
		}
		
		
		queue = queue[1:] // dequeue req
		resqueue = resqueue[1:] // dequeue res
	}
}

// send request to app
func sendReqToApp(w http.ResponseWriter, r *http.Request) {
	
	log.Println("GET request pass to localhost:3000")
	link := "http://localhost:3000"
	// parse url
	url, _ := url.Parse(link)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	//fmt.Println(proxy)
	// serveHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w, r)

}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if reconsile {
		// cluster  not reconsile. request send to  pod
		sendReqToApp(w, r)
	} else {
		// cluster is reconsile. request store in the queue 
		storeReqInQueue(w, r)
	}

}

func main() {

	fmt.Println("Server is up!.")
	h := &Handler{} // call the serverHTTP
	// response msg
	http.Handle("/", http.TimeoutHandler(h, 65*time.Second, "500, Internal error"))
	
	// cluster is reconsile . store the request in the queue
	h1 := func (_ http.ResponseWriter, _ *http.Request)  {
		reconsile = false
		fmt.Println("Cluster is Reconsile : ",reconsile)
	}
	http.HandleFunc("/cluster_Reconsile_Enable",h1)

	// cluster reconsile . working properly.
	h2 := func (res http.ResponseWriter, req *http.Request)  {
		reconsile = true
		fmt.Println("Cluster Reconsile Finished : " ,reconsile)
		popReqQueue(res,req)
	}
	http.HandleFunc("/cluster_Reconsile_Disable",h2)

	// listen server port 50000
	http.ListenAndServe(":50000", nil)

}
