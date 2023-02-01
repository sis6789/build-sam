package build_sam

import (
	"regexp"

	. "github.com/sis6789/global_definition/align"
)

var chrLength = []int{
	0,
	248956422, 242193529, 198295559, 190214555, 181538259,
	170805979, 159345973, 145138636, 138394717, 133797422,
	135086622, 133275309, 114364328, 107043718, 101991189,
	90338345, 83257441, 80373285, 58617616, 64444167,
	46709983, 50818468, 156040895, 57227415,
}

var chrName = []string{
	"",
	"chr1", "chr2", "chr3", "chr4", "chr5",
	"chr6", "chr7", "chr8", "chr9", "chr10",
	"chr11", "chr12", "chr13", "chr14", "chr15",
	"chr16", "chr17", "chr18", "chr19", "chr20",
	"chr21", "chr22", "chrX", "chrY",
}

type QueryName func(r FmdsRaw) string

type event struct {
	chr    int
	start  int
	seq    string
	action byte
}
type matchBlock struct {
	rec    FmdsRaw
	start  int
	tlen   int
	cigar  string
	read   []byte
	human  []byte
	insert []string
	delete []byte
	mdZ    string
	seq    []byte
}

//type samSortType struct {
//	address int
//	line    string
//}

var splitBpSequence = regexp.MustCompile(`([ACGT-]+| +)`)
var insertExp = regexp.MustCompile(`(\d+)i([ACGT])`)
var lastDigits = regexp.MustCompile(`^(\d*[ACGT^]+)*(\d+)$`)
var firstDigits = regexp.MustCompile(`^(\d+)([ACGT^]+\d*)*$`)
