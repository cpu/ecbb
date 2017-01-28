package main

import (
	"flag"
	"fmt"
	"net/http"
)

const greetz = `
     ..............
     .  *      *  .
    [.      '     .]
     .            .
     .     ~~     .
     ..............

Bzztt. Greetings HUMAN. I am

   E L E C T R O N I C
      C O D E B O O K
          B O T

    At your service
`

// main starts a HTTP server on the provided -listen address
func main() {
	listenArg := flag.String("listen", "localhost:6969", "Bind address/port for HTTP server")
	fmt.Printf("%s\n", greetz)
	flag.Parse()

	// TODO(@cpu): Set some timeouts/limits for the HTTP server
	http.HandleFunc("/new", newECB)
	http.ListenAndServe(*listenArg, nil)
}
