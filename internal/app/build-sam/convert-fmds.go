package build_sam

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	. "github.com/sis6789/global_definition/align"
)

var ChanFmds = make(chan FmdsRaw, 1000)
var WgChanFmds sync.WaitGroup
var ChanMB = make(chan *matchBlock, 1000)
var WgChanMB sync.WaitGroup
var matchBlocks []*matchBlock

func convertFmdsToSam() {
	for rec := range ChanFmds {
		mb := convertRecToSamFields(rec)
		ChanMB <- mb
	}
	WgChanFmds.Done()
}

func mergeMB() {
	for mb := range ChanMB {
		if mb != nil {
			matchBlocks = append(matchBlocks, mb)
		}
	}
	// sort
	start := time.Now()
	log.Printf("start sort")
	sort.Slice(matchBlocks, func(i, j int) bool {
		iKey := fmt.Sprintf("%02d%010d%-30s%-30s", matchBlocks[i].rec.Chr, matchBlocks[i].start, matchBlocks[i].rec.FileName, matchBlocks[i].rec.Molecular)
		jKey := fmt.Sprintf("%02d%010d%-30s%-30s", matchBlocks[j].rec.Chr, matchBlocks[j].start, matchBlocks[j].rec.FileName, matchBlocks[j].rec.Molecular)
		return iKey < jKey
	})
	log.Printf("sort %d matchBlocks for %v", len(matchBlocks), time.Since(start))

	WgChanMB.Done()
}

func StartConverter() {
	WgChanMB.Add(1)
	go mergeMB()
	for ix := 0; ix < 10; ix++ {
		WgChanFmds.Add(1)
		go convertFmdsToSam()
	}
}
