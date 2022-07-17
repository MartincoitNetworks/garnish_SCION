Garnish - simple varnish implementation written in Go
===

Garnish is an example of a simple varnish implementation which should demonstrate.

It supports only `Cache-Control` header and keeps all the data in memory. It means that after restart of the application all data will disappear.

The project was described in more details in [the article](https://developer20.com/garnish-simple-varnish-in-go/).

run tests using command-line (you need first cd to the garnish_SCION/garnish/): go test -v garnish_test.go

In the server side, run `go run main.go -listen 127.0.0.1:1234`
In the client side, run `go run main.go -remote 18-ffaa:1:fc1,[127.0.0.1]:1234`

