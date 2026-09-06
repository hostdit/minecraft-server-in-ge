package main

import (
	"flag"
	"log"

	"gemc/internal/mcserver"
)

func main() {
	mcAddr := flag.String("mc", ":25565", "Minecraft listen address")
	httpAddr := flag.String("http", "127.0.0.1:8080", "KML HTTP listen address")
	publicURL := flag.String("url", "http://127.0.0.1:8080", "URL Google Earth uses to reach the KML endpoint")
	flag.Parse()

	log.SetFlags(log.Ltime)
	w := mcserver.NewWorld()
	log.Fatal(mcserver.Run(*mcAddr, *httpAddr, *publicURL, w))
}
