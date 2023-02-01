package main

import (
	"log"
	"testing"
)

func Test_unused(t *testing.T) {
	// remove unused warning
	unused()
	log.Printf("%v", *flagDb)
	log.Printf("%v", *flagGenome)
	log.Printf("%v", *flagMolecular)
	log.Printf("%v", *flagHost)
	log.Printf("%v", *flagOutput)
}
