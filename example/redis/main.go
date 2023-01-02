// Package main contains an example to generate HFID with a Redis KV store
package main

import (
	"context"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"gitlab.com/alielgamal/hfid"
	hfidredis "gitlab.com/alielgamal/hfid/redis"
	"log"
	"time"
)

func main() {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatal(err)
	}
	uc := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{mr.Addr()}})
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
		newHFID, err := hfid.HFID(context.Background(), *g, hfidredis.GeneratorStore{UniversalClient: uc})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(newHFID)
	}
}
