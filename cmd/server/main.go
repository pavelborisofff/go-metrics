package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pavelborisofff/go-metrics/internal/routers"
)

const serverAddrDef = "localhost:8080"

var ServerAddr string

func ParseFlags() {
	var serverAddrFlag string
	flag.StringVar(&serverAddrFlag, "a", serverAddrDef, "Server address")
	flag.Parse()

	serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	if exists {
		serverAddrFlag = serverAddrEnv
	}

	ServerAddr = serverAddrFlag
	msg := fmt.Sprintf("\nServer address: %s", serverAddrFlag)
	log.Println(msg)
}

func main() {
	ParseFlags()
	r := routers.InitRouter()
	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
