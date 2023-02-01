package build_sam

import (
	"io"
	"log"
	"os"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
)

func MakeBamIndex(bamFile *os.File, fn string) {

	var r *sam.Record
	var err error
	var bai bam.Index
	br, err := bam.NewReader(bamFile, 1)
	baiFile, err := os.Create(fn + ".bai")
	if err != nil {
		log.Fatalf("%v", err)
	}

	addIndex := func(r *sam.Record, br *bam.Reader) (isFailed bool) {
		defer func() {
			// ignore panic
			// 패닉이 해당 자료의 열람등에는 문제가 없어 무시함
			if x := recover(); x != nil {
				//log.Printf("%v at %v", x, r.Name)
				isFailed = true
			}
		}()
		err = bai.Add(r, br.LastChunk())
		if err != nil {
			log.Printf("%v %v", r.Name, err)
		}
		return
	}
	var panicCount int
	for {
		r, err = br.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v", err)
		}
		if addIndex(r, br) {
			panicCount++
		}
	}

	err = bam.WriteIndex(baiFile, &bai)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = baiFile.Close()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return
}
