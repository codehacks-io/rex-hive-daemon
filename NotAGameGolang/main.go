package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	signature := getEnvVarSignature()
	fmt.Println(fmt.Sprintf("%sHello from golang", validateSignature(&signature)))
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
		msgBefore := fmt.Sprintf("%s%d of %d: will sleep for %d seconds...", validateSignature(&signature), i+1, total, n)
		if n < 0 {
			_, _ = fmt.Fprintln(os.Stderr, msgBefore)
		} else {
			fmt.Println(msgBefore)
		}

		time.Sleep(time.Duration(secondsToSleep) * time.Second)

		// Message after sleep
		msgAfter := fmt.Sprintf("%s%d of %d: did sleep for %d seconds", validateSignature(&signature), i+1, total, n)
		if n < 0 {
			_, _ = fmt.Fprintln(os.Stderr, msgAfter)
		} else {
			fmt.Println(msgAfter)
		}
	}

	fmt.Println(fmt.Sprintf("%sGood bye from golang", validateSignature(&signature)))
}

func validateSignature(originalSignature *string) string {
	updatedSignature := getEnvVarSignature()
	if *originalSignature != updatedSignature {
		printErrorLn("ERROR IN SIGNATURES: original ", *originalSignature, " does not mach updated one: ", updatedSignature)
	}
	sig := fmt.Sprintf("(%s===%s) ", *originalSignature, updatedSignature)
	return sig
}

func getEnvVarSignature() string {
	pubConn := os.Getenv("REX_PUBLIC_CONNECTIONS")
	privConn := os.Getenv("REX_PRIVATE_CONNECTIONS")
	if pubConn != privConn {
		printErrorLn("CROSS ENV CONTAMINATION", pubConn, privConn)
	}
	return fmt.Sprintf("+%s,-%s", pubConn, privConn)
}

func printErrorLn(msg ...any) {
	colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", 41, "[error]")
	msg = append([]any{colored}, msg...)
	fmt.Println(msg...)
}
