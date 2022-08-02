package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"errors"
	"net/http"
	"net/url"
	"github.com/lucas-clemente/quic-go"
	"io/ioutil"
	"inet.af/netaddr"
	"github.com/bkielbasa/garnish/garnish"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsec-ethz/scion-apps/pkg/quicutil"
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

// create work session
func workSession(session quic.Session) error {
	for {
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			return err
		}
		defer stream.Close()
		data, err := ioutil.ReadAll(stream)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", data)
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
		_, err = stream.Write([]byte(msg))
		if err != nil {
			return err
		}
		_, err = stream.Write(data)
		if err != nil {
			return err
		}
		stream.Close()
	}
}
//change into scion
func runServer(listen netaddr.IPPort) {
	//conn, err := pan.ListenUDP(context.Background(), listen, nil)
	tlsCfg := &tls.Config{
		Certificates: quicutil.MustGenerateSelfSignedCert(),
		NextProtos:   []string{"hello-quic"},
	}
	
	listener, err := pan.ListenQUIC(context.Background(), listen, nil, tlsCfg, nil)
	if err != nil {
		fmt.Printf("listen error")
	}
	defer listener.Close()
	//18-ffaa:1:f53,127.0.0.1:1234
	fmt.Println(listener.Addr())
	for {
		session, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Printf("listener accept error")
		}
		fmt.Println("New session", session.RemoteAddr())
		go func() {
			err := workSession(session)
			var errApplication *quic.ApplicationError
			if err != nil && !(errors.As(err, &errApplication) && errApplication.ErrorCode == 0) {
				fmt.Println(session.RemoteAddr())
			}
		}()
	}		
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
