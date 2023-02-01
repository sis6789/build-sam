package build_sam

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	. "github.com/sis6789/global_definition/align"
	"github.com/sis6789/global_definition/human_reader"
)

func convertRecToSamFields(rec FmdsRaw) *matchBlock {
	var result []*matchBlock
	tk0 := splitBpSequence.FindAllStringSubmatch(rec.Sq, -1)
	var events []event
	if tk0 != nil {
		p := rec.G1
		for _, tk1 := range tk0 {
			switch tk1[0][0] {
			case ' ':
				// 서열 사이의 빈 공간
				if p > rec.G1 {
					events = append(events, event{
						chr:    rec.Chr,
						start:  p,
						seq:    tk1[1],
						action: '9',
					})
				}
				p += len(tk1[1])
			default:
				events = append(events, event{
					chr:    rec.Chr,
					start:  p,
					seq:    tk1[1],
					action: '0',
				})
				p += len(tk1[1])
			}
		}
		// insert variation을 event로 생성
		if rec.Modify != "" {
			iToken := insertExp.FindAllStringSubmatch(rec.Modify, -1)
			for ix := len(iToken) - 1; ix >= 0; ix-- {
				address, err := strconv.Atoi(iToken[ix][1])
				if err != nil {
					log.Fatalf("%v", err)
				}
				events = append(events, event{
					start:  address,
					seq:    iToken[ix][2],
					action: '1',
				})
			}
		}
		// 주소 기준으로 정렬
		sort.Slice(events, func(i, j int) bool {
			switch {
			case events[i].start < events[j].start:
				return true
			case events[i].start > events[j].start:
				return false
			default:
				return events[i].action < events[j].action
			}
		})
		// merge events
		{
			var mb *matchBlock
			var skipCount int
			for _, event := range events {
				switch event.action {
				case '0':
					// sequence
					if mb != nil {
						mb.merge()
						result = append(result, mb)
					}
					mb = new(matchBlock)
					mb.rec = rec
					mb.initialize(event)
				case '1':
					// insert
					if mb != nil {
						mb.insertBP(event)
					}
				case '9':
					// gap
					if mb != nil {
						mb.merge()
						skipCount = len(event.seq)
						mb.cigar += fmt.Sprintf("%dN", skipCount)
						result = append(result, mb)
						mb = nil
					}
				}
			}
			if mb != nil {
				mb.merge()
				result = append(result, mb)
			}
		}
	}

	// 모든 부분을 통합한다.
	for ix := 1; ix < len(result); ix++ {
		result[0].cigar += result[ix].cigar
		tkLast := lastDigits.FindStringSubmatch(result[0].mdZ)
		tkFirst := firstDigits.FindStringSubmatch(result[ix].mdZ)
		if tkLast == nil {
			result[0].mdZ += result[ix].mdZ
		} else if tkFirst == nil {
			result[0].mdZ += result[ix].mdZ
		} else {
			len1, _ := strconv.Atoi(tkLast[2])
			len2, _ := strconv.Atoi(tkFirst[1])
			lenS := strconv.Itoa(len1 + len2)
			result[0].mdZ = result[0].mdZ[:len(result[0].mdZ)-len(tkLast[2])] +
				lenS +
				result[ix].mdZ[len(tkFirst[1]):]
		}
		result[0].seq = append(result[0].seq, result[ix].seq...)
	}
	if len(result) == 0 {
		return nil
	} else {
		return result[0]
	}
}

func (mb *matchBlock) initialize(event event) {
	mb.start = event.start
	mb.read = []byte(event.seq)
	humanString := human_reader.ReadHumanGenome(event.chr, event.start, event.start+len(event.seq)-1)
	mb.human = []byte(strings.ToUpper(humanString))
	mb.insert = make([]string, len(mb.read))
	mb.delete = make([]byte, len(mb.read))
}

func (mb *matchBlock) insertBP(action event) {
	if len(mb.read) == 0 {
		return
	}
	pos := action.start - mb.start
	if pos >= len(mb.read) {
		return
	}
	mb.insert[pos] += action.seq
}

func (mb *matchBlock) merge() {
	if mb == nil {
		return
	}
	if len(mb.read) == 0 {
		return
	}
	deleteActive := 0
	matchActive := 0
	differActive := 0
	applyMatchDiffer := func() {
		if matchActive > 0 {
			mb.cigar += fmt.Sprintf("%dM", matchActive)
			mb.mdZ += fmt.Sprintf("%d", matchActive)
			matchActive = 0
		}
		if differActive > 0 {
			mb.cigar += fmt.Sprintf("%dS", differActive)
			differActive = 0
		}
	}
	for ix := 0; ix < len(mb.read); ix++ {
		if mb.insert[ix] != "" {
			// 삽입
			applyMatchDiffer()
			mb.cigar += fmt.Sprintf("%dI", len(mb.insert[ix]))
			insertBytes := []byte(mb.insert[ix])
			mb.seq = append(mb.seq, insertBytes...)
		}
		if mb.read[ix] == '-' {
			// 누락
			applyMatchDiffer()
			if deleteActive == 0 {
				mb.mdZ += "^"
			}
			mb.mdZ += string(mb.human[ix])
			deleteActive++
			continue
		}
		// 이전 누락을 종료
		if deleteActive > 0 {
			mb.cigar += fmt.Sprintf("%dD", deleteActive)
			deleteActive = 0
		}
		if mb.human[ix] == mb.read[ix] {
			if differActive > 0 {
				// 이전 치환을 종료
				mb.cigar += fmt.Sprintf("%dS", differActive)
				differActive = 0
			}
			matchActive++
		} else {
			if matchActive > 0 {
				// 이전 매치를 종료
				mb.cigar += fmt.Sprintf("%dM", matchActive)
				mb.mdZ += fmt.Sprintf("%d", matchActive)
				matchActive = 0
			}
			mb.mdZ += string(mb.human[ix])
			differActive++
		}
		mb.seq = append(mb.seq, mb.read[ix])
	}
	applyMatchDiffer()
}
