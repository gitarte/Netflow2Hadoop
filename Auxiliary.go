package main

import "time"

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
