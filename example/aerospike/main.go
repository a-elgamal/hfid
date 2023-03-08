// Package main contains an example to generate HFID with a Redis KV store
package main

import (
	"context"
	"fmt"
	aero "github.com/aerospike/aerospike-client-go/v6"
	"gitlab.com/alielgamal/hfid"
	hfidaero "gitlab.com/alielgamal/hfid/aerospike"
	"log"
	"time"
)

func main() {
	c, aeroErr := aero.NewClient("localhost", 3000)
	if aeroErr != nil {
		log.Fatal(aeroErr)
	}

	g, err := hfid.NewGenerator("Example", "E-", hfid.DefaultEncoding, 1, 1)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	defer func() {
		fmt.Printf("\n Execution Took %dms\n", time.Now().UnixMilli()-start.UnixMilli())
	}()
	hfidCount := 1000
	fmt.Printf("Generating %d HFIDs:\n", hfidCount)
	for i := 0; i < hfidCount; i++ {
		newHFID, err := hfid.HFID(context.Background(), *g, hfidaero.GeneratorStore{Client: c, Namespace: "test", Set: "hfid"})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(newHFID)
	}
}
