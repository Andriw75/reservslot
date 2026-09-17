package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	fmt.Printf("JWT_SECRET=%s\n", hex.EncodeToString(b))
}
