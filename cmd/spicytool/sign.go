package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"filippo.io/torchwood"
	"github.com/transparency-dev/tessera"
	"github.com/warpfork/spicytool"
	"github.com/warpfork/spicytool/contexts"
)

var (
	signFlags  = flag.NewFlagSet("sign", flag.ContinueOnError)
	logPath    = signFlags.String("logdir", "tlog", "directory for transparency log")
	filename   = signFlags.String("f", "", "File to sign (use \"-\" for stdin)")
	hint       = signFlags.String("n", "", "Filename / context hint (defaults to filename)")
	outputFile = signFlags.String("o", "", "Output file for signature (defaults to stdout)")
)

func sign(args []string) {
	// Flag parsing.
	signFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: spicytool sign [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		signFlags.PrintDefaults()
	}
	if err := signFlags.Parse(args); err != nil {
		os.Exit(exitFlagMisuse)
	}
	if filename == nil || *filename == "" {
		fmt.Fprintf(os.Stderr, "Usage error: '-f' flag is not optional\n")
		os.Exit(exitFlagMisuse)
	}
	if os.Getenv("LOG_KEY") == "" {
		fmt.Fprintf(os.Stderr, "Usage error: LOG_KEY env var must be provided and contain a private key for signing tlog checkpoints\n")
		os.Exit(exitFlagMisuse)
	}

	// Tlog setup.
	ctx := context.Background()
	lh, err := spicytool.OperateLog(ctx,
		*logPath,
		os.Getenv("LOG_KEY"),
		tessera.NewWitnessGroup(0)) // TODO replace with more energetic defaults
	if err != nil {
		log.Fatalln("failed to open log for appending:", err)
	}
	defer lh.Shutdown(ctx)

	// Open output file.
	// If this would fail, we want to fail before doing any of the other heavy lifting.
	output := os.Stdout
	if outputFile != nil && *outputFile != "" {
		output, err = os.OpenFile(*outputFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			log.Fatalln("failed to open output file:", err)
		}
	}

	// Compute our entry.
	subject := os.Stdin
	if *filename != "-" {
		subject, err = os.Open(*filename)
		if err != nil {
			log.Fatalln("could not open file:", err)
		}
	}
	var hintActual string
	if hint == nil || *hint == "" {
		hintActual = *filename
	} else {
		hintActual = *hint
	}
	entry, err := contexts.RecordForBody(subject, hintActual)
	if err != nil {
		log.Fatalln("error while streaming the subject:", err)
	}

	// Append log.
	idx, err := lh.AppendAndAwait(ctx, entry)
	if err != nil {
		log.Fatalln("failed to append to transparency log:", err)
	}

	// Generate inclusion proof for the entry.
	recordProof, checkpointRaw, err := lh.GenerateProof(ctx, int64(idx))
	if err != nil {
		log.Fatalln("failed to generate inclusion proof:", err)
	}

	// Emit spicysig.
	var spicybytes []byte
	if hintActual != "" {
		spicybytes = torchwood.FormatProofWithExtraData(int64(idx), []byte(hintActual), recordProof, checkpointRaw)
	} else {
		spicybytes = torchwood.FormatProof(int64(idx), recordProof, checkpointRaw)
	}
	output.Write(spicybytes)
}
