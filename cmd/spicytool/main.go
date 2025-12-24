package main

import (
	"fmt"
	"os"
)

var (
	exitFlagMisuse = 2
)

func main() {
	// Parse subcommands or bail to usage.
	// (Each subcommand function will then use its own FlagSet.)
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitFlagMisuse)
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "sign":
		sign(os.Args[2:])
	case "verify":
		verify(os.Args[2:])
	case "keygen":
		keygen(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand %q\n\n", subcommand)
		printUsage()
		os.Exit(exitFlagMisuse)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `SpicyTool - Generate and verify Spicy Signatures

Usage:
  spicytool <command> [options]

Commands:
  sign      Generate a spicy signature
  verify    Verify a spicy signature
  keygen    Generate a key useful for running a log and performing sign
  help      Show this help message

Sign options:
`)
	signFlags.PrintDefaults()

	fmt.Fprint(os.Stderr, `
Verify options:
`)
	verifyFlags.PrintDefaults()

	fmt.Fprint(os.Stderr, `
Keygen options:
  <origin-name>
        positional argument: the name to use in the key
`)

	fmt.Fprint(os.Stderr, `
Examples:
  spicytool sign -f document.txt -o document.txt.sig
  spicytool sign -f document.txt -n "my document" -o document.txt.sig
  spicytool verify -f document.txt -s document.txt.sig
  echo "hello world" | spicytool sign -f -
  echo "hello world" | spicytool verify -f - -s hello.sig

`)
}
