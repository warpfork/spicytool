package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"filippo.io/torchwood"
	"golang.org/x/mod/sumdb/tlog"
)

var (
	verifyFlags    = flag.NewFlagSet("verify", flag.ContinueOnError)
	verifyFilename = verifyFlags.String("f", "", "File to verify (use \"-\" for stdin)")
	signatureFile  = verifyFlags.String("s", "", "Path to file containing the spicy signature")
	verifyHint     = verifyFlags.String("n", "", "Filename / context hint (defaults to filename)")
)

//go:embed default_policy.txt
var defaultWitnessPolicy []byte

func verify(args []string) {
	// Flag and env parsing.
	verifyFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: spicytool verify [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		verifyFlags.PrintDefaults()
	}
	if err := verifyFlags.Parse(args); err != nil {
		os.Exit(exitFlagMisuse)
	}
	if verifyFilename == nil || *verifyFilename == "" {
		fmt.Fprintf(os.Stderr, "Usage error: '-f' flag is not optional\n")
		os.Exit(exitFlagMisuse)
	}
	if signatureFile == nil || *signatureFile == "" {
		fmt.Fprintf(os.Stderr, "Usage error: '-s' flag is not optional\n")
		os.Exit(exitFlagMisuse)
	}
	signatureBytes, err := os.ReadFile(*signatureFile)
	if err != nil {
		log.Fatalln("failed to read signature file:", err)
	}
	witnessPolicyBytes := defaultWitnessPolicy
	if path := os.Getenv("LOG_WITNESS_POLICY"); path != "" {
		var err error
		witnessPolicyBytes, err = os.ReadFile(path)
		if err != nil {
			log.Fatalln("failed to read witness policy file:", err)
		}
	}

	// Reify args.
	// (The ones we can, anyway.  `torchwood.VerifyProof` combines the parse
	// and the verification of the spicysig into one large operation;
	// as a result, we *can't* check for obvious format errors there until
	// after we've all the other heavy lifting including the full IO and hashing of the subject.)
	witnessPolicy, err := torchwood.ParsePolicy(witnessPolicyBytes)
	if err != nil {
		log.Fatalln("failed to parse witness policy:", err)
	}

	// Compute the entry -- same way as sign would've.
	entry, _ := computeEntry(*verifyFilename, *verifyHint)

	// Proof verification.
	// This checks the signatures in the checkpoint section of the spicysig against the policy,
	// and computes that our entry plus the inclusion proof section in the spicysig reproduces the treehash in the checkpoint.
	// It doesn't need any further online information nor any other tlog reads to check this!  :D
	if err := torchwood.VerifyProof(witnessPolicy, tlog.RecordHash(entry), signatureBytes); err != nil {
		log.Fatalln("failed to verify proof:", err)
	}

	fmt.Fprintf(os.Stderr, "verification success!\n")
}
