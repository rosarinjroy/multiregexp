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

	// Fields used for submatch based replacement
	compiledRegexWithSubmatch *regexp.Regexp

	// For brute force method, we will use the compiledRegexList
}

func NewMultiRegexp() *MultiRegexp {
	retVal := &MultiRegexp{}
	retVal.algorithm = MultiRegexpAlgorithmOR
	return retVal
}

func NewMultiRegexpUsingSubmatch() *MultiRegexp {
	return &MultiRegexp{}
}

func NewMultiRegexpUsingBruteForce() *MultiRegexp {
	return &MultiRegexp{}
}

// Adds a new replacement to the list of replacements.
func (p *MultiRegexp) AddReplacement(old, new string) {
	p.fromRegexes = append(p.fromRegexes, old)
	p.toStrings = append(p.toStrings, new)
	p.mustInitialize = true
}

func (p *MultiRegexp) Compile() {
	if !p.mustInitialize {
		return
	}

	switch p.algorithm {
	case MultiRegexpAlgorithmOR:
		p.CompileUsingOR()
	case MultiRegexpAlgorithmSubmatch:
		p.CompileUsingSubmatch()
	case MultiRegexpAlgorithmBruteForce:
		p.CompileUsingBruteForce()
	}
}

func (p *MultiRegexp) CompileUsingOR() {
	if p.compiledRegexWithSubmatch == nil {
		p.compiledRegexWithSubmatch = regexp.MustCompile(strings.Join(p.fromRegexes, "|"))
	}
	for _, old := range p.fromRegexes {
		p.compiledRegexList = append(p.compiledRegexList, regexp.MustCompile(old))
	}
	if p.memoize {
		p.memoizedMatches = make(map[string][]byte)
	}
}

func (p *MultiRegexp) CompileUsingSubmatch() {
	if p.compiledRegex == nil {
		p.compiledRegex = regexp.MustCompile("(" + strings.Join(p.fromRegexes, ")|(") + ")")
	}
}

func (p *MultiRegexp) CompileUsingBruteForce() {
	if p.compiledRegexList == nil {
		for _, old := range p.fromRegexes {
			p.compiledRegexList = append(p.compiledRegexList, regexp.MustCompile(old))
		}
	}
}

func (p *MultiRegexp) PrintStats() {
	fmt.Println("FromRegexes:", len(p.fromRegexes))
	fmt.Println("ToStrings:", len(p.toStrings))
}

func (p *MultiRegexp) ReplaceAll(text []byte) []byte {
	switch p.algorithm {
	case MultiRegexpAlgorithmOR:
		return p.ReplaceAllUsingOR(text)
	case MultiRegexpAlgorithmSubmatch:
		return p.ReplaceAllUsingSubmatch(text)
	case MultiRegexpAlgorithmBruteForce:
		return p.ReplaceAllUsingBruteForce(text)
	}
	panic(fmt.Sprintf("Unknown algorithm: %s", p.algorithm))
}

func (p *MultiRegexp) ReplaceAllUsingOR(text []byte) []byte {
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
		for i := range p.compiledRegexList {
			if p.memoize {
				replacement, ok := p.memoizedMatches[string(text[match[0]:match[1]])]
				if ok {
					retVal = append(retVal, replacement...)
					lastIndex = match[1]
					break
				}
			}

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

func (p *MultiRegexp) ReplaceAllUsingSubmatch(text []byte) []byte {
	retVal := make([]byte, 0, len(text))
	matches := p.compiledRegex.FindAllSubmatchIndex(text, -1)

	lastIndex := 0
	for _, match := range matches {
		// fmt.Println(match)
		for i := 2; i < len(match); i += 2 {
			if match[i] == -1 {
				break
			}
			if lastIndex < match[i] {
				retVal = append(retVal, text[lastIndex:match[i]]...)
			}
			replacement := []byte(p.toStrings[(i-2)/2])
			retVal = append(retVal, replacement...)
			lastIndex = match[i+1]
			break
		}
	}
	retVal = append(retVal, text[lastIndex:]...)
	return retVal
}

func (p *MultiRegexp) ReplaceAllUsingBruteForce(text []byte) []byte {
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
