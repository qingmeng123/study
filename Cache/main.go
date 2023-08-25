package main

import (
	"fmt"
	"study/Cache/DuryunCache"

	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	DuryunCache.NewGroup("scores", 2<<10, DuryunCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := DuryunCache.NewHTTPPool(addr)
	log.Println("DuryunCache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
