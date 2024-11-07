package main

import (
	"log"
	"net"
	"net/http"
)

func checkInternalServerError(err error, w http.ResponseWriter) {
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// used to auto detect the active local IP address - not used yet
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

type ValidateRequire struct {
	requiredChar []rune
	amount       int
}

// returns false if a char in 'blacklist' is found in 's'
// returns false if counted chars from 'require.requiredChar' is not equal to 'require.amount'
func ValidateString(s string, blacklist []rune, requires []ValidateRequire) bool {
	requireCounts := make([]int, len(requires))

	for _, ch := range s {
		for _, bl := range blacklist {
			if ch == bl {
				return false
			}
		}
		for i := 0; i < len(requires); i++ {
			for _, req := range requires[i].requiredChar {
				if ch == req {
					requireCounts[i]++
				}
			}
		}
	}

	for i := 0; i < len(requires); i++ {
		if requireCounts[i] < requires[i].amount {
			return false
		}
	}

	return true
}

func minInt(x int, y int) int {
	if x > y {
		return y
	}
	return x
}
