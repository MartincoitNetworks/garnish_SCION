package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	//"time"

	"inet.af/netaddr"

	"github.com/bkielbasa/garnish/garnish"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
)

func main() {
	var listen pan.IPPortValue
	flag.Var(&listen, "listen", "[Server] local IP:port to listen on")
	remoteAddr := flag.String("remote", "", "[Client] Remote (i.e. the server's) SCION Address (e.g. 17-ffaa:1:1,[127.0.0.1]:12345)")
	flag.Parse()
	if (listen.Get().Port() > 0) == (len(*remoteAddr) > 0) {
		panicOnErr(fmt.Errorf("either specify -listen for server or -remote for client"))
	}

	if listen.Get().Port() > 0 {
		runServer(listen.Get())
	} else {
		runGarnish(*remoteAddr)
	}

}

func runGarnish(address string) {
	u := url.URL{Scheme: "http", Host: "localhost:8088"}
	g := garnish.New(u)
	handlers := func(w http.ResponseWriter, req *http.Request) {
		g.ServeHTTP(w, req, address)
		xcache := w.Header().Get("X-Cache")
		fmt.Println("Real header:" + xcache)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handlers)
	log.Fatal(http.ListenAndServe(":8088", mux))
	fmt.Println(address)

}

// func runClient(address string) {

// 	fmt.Println(address)
// 	u := url.URL{Scheme: "http", Host: "localhost:8088"}
// 	g := garnish.New(u)
// 	expectedXCacheHeaders := []string{garnish.XcacheMiss, garnish.XcacheHit}

// 	for _, expectedHeader := range expectedXCacheHeaders {
// 		req := httptest.NewRequest(http.MethodGet, "http://localhost:8088", nil)
// 		w := httptest.NewRecorder()
// 		g.ServeHTTP(w, req, address)

// 		xcache := w.Header().Get("X-Cache")
// 		fmt.Println("Expected header: " + expectedHeader)
// 		fmt.Println("Real header:" + xcache)

// 	}
// }

//change into scion
func runServer(listen netaddr.IPPort) {
	conn, err := pan.ListenUDP(context.Background(), listen, nil)
	if err != nil {
		fmt.Printf("listen error")
	}
	//defer conn.Close()
	fmt.Print("Hello! ")
	fmt.Println(conn.LocalAddr())
	for true {
		buffer := make([]byte, 1024*16*1024)
		n, from, err := conn.ReadFrom(buffer)
		if err != nil {
			fmt.Printf("read error")
		}
		const msg = `<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>SCION demo</title>
		</head>
		<body>
		<h1>Welcome to Granish</h1>
		<p>Using SCION architecture</p>
		<img src="https://www.cylab.cmu.edu/_files/images/research/scion/scion-banner.png" alt="SCION banner">
		</body>
		</html>`
		//msg := fmt.Sprintf("Aaaaa, i love it")
		n, err = conn.WriteTo([]byte(msg), from)
		if err != nil {
			fmt.Printf("write error")
		}
		fmt.Printf("Wrote %d bytes.\n", n)
		//time.Sleep(time.Millisecond * 30)
	}
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
