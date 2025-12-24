package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"golang.org/x/mod/sumdb/note"
)

func keygen(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: spicytool keygen <origin-name>\n\n")
		os.Exit(exitFlagMisuse)
	}
	origin := args[0]

	skey, vkey, err := note.GenerateKey(rand.Reader, origin)
	if err != nil {
		log.Fatalln("Error generating keys:", err)
	}

	fmt.Printf("Private key (for LOG_KEY when signing): %s\n", skey)
	fmt.Printf("Public key (for 'log' entry in policy file): %s\n", vkey)
}
