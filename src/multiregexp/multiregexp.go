package parallelreplace

import (
	"fmt"
	"regexp"
	"strings"
)

type MultiRegexpAlgorithm string

const (
	MultiRegexpAlgorithmOR         MultiRegexpAlgorithm = "OR"
	MultiRegexpAlgorithmSubmatch   MultiRegexpAlgorithm = "SUBMATCH"
	MultiRegexpAlgorithmBruteForce MultiRegexpAlgorithm = "BRUTE_FORCE"
)

type MultiRegexp struct {
	fromRegexes []string
	toStrings   []string

	algorithm MultiRegexpAlgorithm

	mustInitialize bool

	// Fields used for OR based replacement with optional memoization
	compiledRegex     *regexp.Regexp
	compiledRegexList []*regexp.Regexp
	memoize           bool
	memoizedMatches   map[string][]byte

	// For submatch based replacement, we will use the compiledRegex
	// to store the compiled regex with submatch.

	// For brute force method, we will use the compiledRegexList
}

func NewMultiRegexp() *MultiRegexp {
	retVal := &MultiRegexp{}
	retVal.algorithm = MultiRegexpAlgorithmOR
	retVal.mustInitialize = true
	return retVal
}

func NewMultiRegexpUsingSubmatch() *MultiRegexp {
	retVal := &MultiRegexp{}
	retVal.algorithm = MultiRegexpAlgorithmSubmatch
	retVal.mustInitialize = true
	return retVal
}

func NewMultiRegexpUsingBruteForce() *MultiRegexp {
	retVal := &MultiRegexp{}
	retVal.algorithm = MultiRegexpAlgorithmBruteForce
	retVal.mustInitialize = true
	return retVal
}

func (p *MultiRegexp) GetAlgorithm() MultiRegexpAlgorithm {
	return p.algorithm
}

func (p *MultiRegexp) Memoize(memoize bool) {
	if p.algorithm == MultiRegexpAlgorithmOR {
		p.memoize = memoize
	} else {
		panic(fmt.Sprintf("Memize is only supported for OR algorithm, not %s", p.algorithm))
	}
}

func (p *MultiRegexp) IsMemoized() bool {
	return p.memoize
}

// Adds a new replacement to the list of replacements.
func (p *MultiRegexp) AddReplacement(old, new string) {
	p.fromRegexes = append(p.fromRegexes, old)
	p.toStrings = append(p.toStrings, new)
	p.mustInitialize = true
}

func (p *MultiRegexp) initializeIfNeeded() {
	if !p.mustInitialize {
		return
	}

	switch p.algorithm {
	case MultiRegexpAlgorithmOR:
		p.compileForOR()
	case MultiRegexpAlgorithmSubmatch:
		p.compileForSubmatch()
	case MultiRegexpAlgorithmBruteForce:
		p.compileForBruteForce()
	default:
		panic(fmt.Sprintf("Unknown algorithm: %s", p.algorithm))
	}
	p.mustInitialize = false
}

func (p *MultiRegexp) compileForOR() {
	p.compiledRegex = regexp.MustCompile(strings.Join(p.fromRegexes, "|"))

	for _, old := range p.fromRegexes {
		p.compiledRegexList = append(p.compiledRegexList, regexp.MustCompile(old))
	}
	if p.memoize {
		p.memoizedMatches = make(map[string][]byte)
	}
}

func (p *MultiRegexp) compileForSubmatch() {
	if p.compiledRegex == nil {
		p.compiledRegex = regexp.MustCompile("(" + strings.Join(p.fromRegexes, ")|(") + ")")
	}
}

func (p *MultiRegexp) compileForBruteForce() {
	if p.compiledRegexList == nil {
		for _, old := range p.fromRegexes {
			p.compiledRegexList = append(p.compiledRegexList, regexp.MustCompile(old))
		}
	}
}

// func (p *MultiRegexp) PrintStats() {
// 	fmt.Println("FromRegexes:", len(p.fromRegexes))
// 	fmt.Println("ToStrings:", len(p.toStrings))
// }

func (p *MultiRegexp) ReplaceAll(text []byte) []byte {
	p.initializeIfNeeded()

	switch p.algorithm {
	case MultiRegexpAlgorithmOR:
		return p.replaceAllUsingOR(text)
	case MultiRegexpAlgorithmSubmatch:
		return p.replaceAllUsingSubmatch(text)
	case MultiRegexpAlgorithmBruteForce:
		return p.replaceAllUsingBruteForce(text)
	}
	panic(fmt.Sprintf("Unknown algorithm: %s", p.algorithm))
}

func (p *MultiRegexp) replaceAllUsingOR(text []byte) []byte {
	retVal := make([]byte, 0, len(text))
	matches := p.compiledRegex.FindAllSubmatchIndex(text, -1)
	// fmt.Println(re.NumSubexp())
	lastIndex := 0
	for _, match := range matches {
		// Each match is of length 2. Those two elements are the start and end of
		// the full match.
		if lastIndex < match[0] {
			retVal = append(retVal, text[lastIndex:match[0]]...)
		}
		if p.memoize {
			replacement, ok := p.memoizedMatches[string(text[match[0]:match[1]])]
			if ok {
				retVal = append(retVal, replacement...)
				lastIndex = match[1]
				continue
			}
		}

		for i := range p.compiledRegexList {
			if p.compiledRegexList[i].Match(text[match[0]:match[1]]) {
				replacement := []byte(p.toStrings[i])
				retVal = append(retVal, replacement...)
				lastIndex = match[1]
				if p.memoize {
					p.memoizedMatches[string(text[match[0]:match[1]])] = replacement
				}
				break
			}
		}
	}
	retVal = append(retVal, text[lastIndex:]...)
	return retVal
}

func (p *MultiRegexp) replaceAllUsingSubmatch(text []byte) []byte {
	retVal := make([]byte, 0, len(text))
	matches := p.compiledRegex.FindAllSubmatchIndex(text, -1)

	lastIndex := 0
	for _, match := range matches {
		// fmt.Println(match)
		for i := 2; i < len(match); i += 2 {
			if match[i] == -1 {
				continue
			}
			if lastIndex < match[i] {
				retVal = append(retVal, text[lastIndex:match[i]]...)
			}
			replacement := []byte(p.toStrings[(i-2)/2])
			retVal = append(retVal, replacement...)
			lastIndex = match[i+1]
		}
	}
	retVal = append(retVal, text[lastIndex:]...)
	return retVal
}

func (p *MultiRegexp) replaceAllUsingBruteForce(text []byte) []byte {
	var retVal []byte

	for i := range p.compiledRegexList {
		retVal = make([]byte, 0, len(text))
		matches := p.compiledRegexList[i].FindAllSubmatchIndex(text, -1)
		replacement := []byte(p.toStrings[i])

		lastIndex := 0

		for _, match := range matches {
			if lastIndex < match[0] {
				retVal = append(retVal, text[lastIndex:match[0]]...)
			}
			retVal = append(retVal, replacement...)
			lastIndex = match[1]
		}
		retVal = append(retVal, text[lastIndex:]...)

		text = retVal
	}

	return text
}
