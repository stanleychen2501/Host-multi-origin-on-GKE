//cn326.cn
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/sebest/xff"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ips := strings.Split(r.Header.Get("X-Forwarded-For"))
		xff := strings.Split(r.Header.Get("X-Forwarded-For"), ", ")
		log.Printf("xff: %+v", xff)
		w.Write([]byte("A website , XFF IP is " + strings.Join(xff, ", ") + "\n"))
	})

	xffmw, _ := xff.Default()
	http.ListenAndServe(":8080", xffmw.Handler(handler))
}
