package main

import (
	"fmt"
	"os"
	"time"
)

// GetTimestamp Helper function to decode timestamp in NetFlow v5 payload
func GetTimestamp(sec []byte, nsec []byte) string {
	//	seconds part
	s := int64(sec[0])*256*256*256 +
		int64(sec[1])*256*256 +
		int64(sec[2])*256 +
		int64(sec[3])

	//	nano seconds part
	n := int64(nsec[0])*256*256*256 +
		int64(nsec[1])*256*256 +
		int64(nsec[2])*256 +
		int64(nsec[3])
	t := time.Unix(s, n)
	return t.Format("2006-01-02 15:04:05.000000000")
}

// RecoverAnyPanic saves program from unexpected crushes
func RecoverAnyPanic(where string) {
	if r := recover(); r != nil {
		fmt.Printf("Recovered in %s with message %s\n", where, r)
	}
}

// ExitOnError stop execution becouse of fatal error
func ExitOnError(where string, err error) {
	fmt.Fprintf(os.Stderr, "Fatal error in %s: %v\n", where, err)
	os.Exit(1)
}

// LogOnError prints message and error if given
func LogOnError(where string, err error) {
	fmt.Fprintf(os.Stderr, "Error in %s: %v\n", where, err)
}
