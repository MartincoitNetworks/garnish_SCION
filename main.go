package main

import (
	"context"
	//"errors"
	"flag"
	"fmt"
	"inet.af/netaddr"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
	//"os"

	"github.com/bkielbasa/garnish/garnish"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	//"garnish_SCION/garnish/garnish"
)

func main() {
	var listen pan.IPPortValue
	// err := listen.Set("127.0.0.1:1234")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	flag.Var(&listen, "listen", "[Server] local IP:port to listen on")
	remoteAddr := flag.String("remote", "", "[Client] Remote (i.e. the server's) SCION Address (e.g. 17-ffaa:1:1,[127.0.0.1]:12345)")
	flag.Parse()
	if (listen.Get().Port() > 0) == (len(*remoteAddr) > 0) {
		panicOnErr(fmt.Errorf("either specify -listen for server or -remote for client"))
	}

	if listen.Get().Port() > 0 {
		runServer(listen.Get())
	} else {
		runClient(*remoteAddr)
	}

	// stop := runServer(listen.Get())
	// defer stop() //stop will run when we're finished

}
func runClient(address string) {

	fmt.Println(address)
	g := garnish.New(url.URL{Scheme: "http", Host: "localhost:8088"})
	expectedXCacheHeaders := []string{garnish.XcacheMiss, garnish.XcacheHit}

	for _, expectedHeader := range expectedXCacheHeaders {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:8088", nil)
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req, address)

		xcache := w.Header().Get("X-Cache")
		fmt.Println("Expected header: " + expectedHeader)
		fmt.Println("Real header:" + xcache)
	}
}

//change into scion
func runServer(listen netaddr.IPPort) func() {
	conn, err := pan.ListenUDP(context.Background(), listen, nil)
	if err != nil {
		fmt.Printf("listen error")
	}
	defer conn.Close()
	fmt.Print("Hello! ")
	fmt.Println(conn.LocalAddr())
	buffer := make([]byte, 16*1024)
	n, from, err := conn.ReadFrom(buffer)
	if err != nil {
		fmt.Printf("read error")
	}
	msg := fmt.Sprintf("This is the data in server!")
	n, err = conn.WriteTo([]byte(msg), from)
	if err != nil {
		fmt.Printf("write error")
	}
	fmt.Printf("Wrote %d bytes.\n", n)
	time.Sleep(time.Millisecond * 30)
	return func() {
		panicOnErr(err)
	}
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
