package build_sam

import (
	"log"
	"time"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
)

func SaveSam(bioSam *sam.Writer, bioBam *bam.Writer, qn QueryName) {
	start := time.Now()
	log.Printf("start sam/bam write")
	for _, mb := range matchBlocks {
		mb.rec.Name = qn(mb.rec)
		writeSamLine(bioSam, mb.rec, mb)
		writeSamLine(bioBam, mb.rec, mb)
	}
	log.Printf("sam/bam write for %v", time.Since(start))
}
