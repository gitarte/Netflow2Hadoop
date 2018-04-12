package main

import (
	"log"
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
		log.Printf("Recovered in %s with message %s\n", where, r)
	}
}

// ExitOnError stop execution because of fatal error
func ExitOnError(where string, err error) {
	log.Fatalf("Fatal error in %s: %v\n", where, err)
}

// LogOnError prints message and error if given
func LogOnError(where string, err error) {
	log.Printf("Error in %s: %v\n", where, err)
}
