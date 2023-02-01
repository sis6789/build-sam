package build_sam

import (
	"log"
	"os"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"

	. "github.com/sis6789/global_definition/align"
)

var systemRefs []*sam.Reference
var header *sam.Header

func init() {
	var err error
	for ix := 1; ix < len(chrName); ix++ {
		ref, err := sam.NewReference(chrName[ix], "", "", chrLength[ix], nil, nil)
		if err != nil {
			log.Fatalf("%v", err)
		}
		systemRefs = append(systemRefs, ref)
	}
	if header, err = sam.NewHeader(nil, systemRefs); err != nil {
		log.Fatalf("%v", err)
	}
	header.Version = "1.6"
	header.SortOrder = sam.Coordinate
}

func PrepareSam(file *os.File) *sam.Writer {
	samWriter, err := sam.NewWriter(file, header, sam.FlagDecimal)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return samWriter
}

func PrepareBam(file *os.File) *bam.Writer {
	bamWriter, err := bam.NewWriterLevel(file, header, 9, 1)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return bamWriter
}

func writeSamLine(bs interface{}, thisRec FmdsRaw, match *matchBlock) {

	var err error

	var samRec sam.Record
	samRec.Name = thisRec.Name
	if thisRec.Direction == "m" {
		samRec.Flags += sam.Reverse
	}
	samRec.MapQ = 255
	samRec.Ref = systemRefs[thisRec.Chr-1]
	samRec.Pos = match.start - 1
	samRec.Cigar, err = sam.ParseCigar([]byte(match.cigar))
	if err != nil {
		log.Fatalf("%v", err)
	}
	samRec.Seq = sam.NewSeq(match.seq)
	samRec.TempLen = 0 // len(match.seq)
	md, err := sam.NewAux(sam.NewTag("MD"), match.mdZ)
	if err != nil {
		log.Fatalf("%v", err)
	}
	samRec.AuxFields = []sam.Aux{md}

	switch bs.(type) {
	case *sam.Writer:
		err = bs.(*sam.Writer).Write(&samRec)
		if err != nil {
			log.Fatalf("%v", err)
		}
	case *bam.Writer:
		err = bs.(*bam.Writer).Write(&samRec)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	return
}
