package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	fmt.Println("Hello from golang")
	flag.Parse()
	args := flag.Args()

	total := len(args)
	for i, arg := range args {
		n, err := strconv.Atoi(arg)

		secondsToSleep := n
		if secondsToSleep < 0 {
			secondsToSleep = secondsToSleep * -1
		}

		if err != nil {
			panic(err)
		}

		// Message before sleep
		msgBefore := fmt.Sprintf("%d of %d: will sleep for %d seconds...", i, total, n)
		if n < 0 {
			_, _ = fmt.Fprintln(os.Stderr, msgBefore)
		} else {
			fmt.Println(msgBefore)
		}

		time.Sleep(time.Duration(secondsToSleep) * time.Second)

		// Message after sleep
		msgAfter := fmt.Sprintf("%d of %d: did sleep for %d seconds", i, total, n)
		if n < 0 {
			_, _ = fmt.Fprintln(os.Stderr, msgAfter)
		} else {
			fmt.Println(msgAfter)
		}
	}

	fmt.Println("Good bye from golang")
}
