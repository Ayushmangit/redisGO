package main

import (
	"fmt"
	"time"

	"github.com/Ayushmangit/redisGO/store/kv"
)

// func main() {
// 	cmds := []Command{
// 		{OP: "SET", Key: "env", Value: "development"},
// 		{OP: "SET", Key: "version", Value: "0.0.1"},
// 		{OP: "SET", Key: "debug", Value: "true"},
// 		{OP: "GET", Key: "env"},
// 		{OP: "SET", Key: "region", Value: "asia-south"},
// 		{OP: "GET", Key: "version"},
// 	}
//
// 	s := kv.NewStore(int(time.Second))
// 	RestoreOnBoot(s, cmds)
//
// 	fmt.Println("restored, store size:", s.Len())
// 	fmt.Println("Keys:", s.Keys())
// }

func main() {
	s := kv.NewStore(20, 5*time.Second)

	s.Set("pageviews", "0", 0)
	const hits = 1000

	cmds := make([]Command, hits)
	for i := range cmds {
		cmds[i] = Command{OP: "INCR", Key: "pageviews"}
	}

	cmds = append(cmds,
		Command{OP: "INCR", Key: ""},
		Command{OP: "WAT", Key: "pageviews"},
	)

	RestoreOnBoot(s, cmds)
	got, _ := s.Get("pageviews")
	fmt.Printf("expected, %d\n", hits)
	fmt.Printf("got: %s\n", got)
}
