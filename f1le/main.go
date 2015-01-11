package f1le

import (
	"log"
	"math/rand"
	"os"
	"time"
)

func Main() {
	rand.Seed(time.Now().UTC().UnixNano())
	
	if len(os.Args) != 3 {
		log.Fatal("Usage: ", os.Args[0], " <port> <root path>")
	}

	// Setup configuration
	configuration, err := LoadConfig(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	// Serve
	if err := Serve(os.Args[1], configuration); err != nil {
		log.Fatal(err)
	}
}
