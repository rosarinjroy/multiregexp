package parallelreplace

import (
	"fmt"
	"regexp"
	"strings"
	"time"
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
	compiledRegex                     *regexp.Regexp
	compiledRegexList                 []*regexp.Regexp
	memoize                           bool
	memoizedMatches                   map[string][]byte
	compiledMemoizedMatchesRegex      *regexp.Regexp
	mustRecompileMemoizedMatchesRegex bool

	// For submatch based replacement, we will use the compiledRegex
	// to store the compiled regex with submatch.

	// For brute force method, we will use the compiledRegexList
}

type MultiRegexpStats struct {
	numReplacements     int
	numMatches          int
	numMemoizedMatches  int
	numMemoizedEntries  int
	regexpMatchDuration time.Duration
}

func NewMultiRegexp() *MultiRegexp {
	retVal := &MultiRegexp{}
	retVal.algorithm = MultiRegexpAlgorithmOR
	retVal.mustInitialize = true
	retVal.mustRecompileMemoizedMatchesRegex = true
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
		panic(fmt.Sprintf("Memoize is only supported for OR algorithm, not %s", p.algorithm))
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

func (p *MultiRegexp) ReplaceAll(text []byte) ([]byte, MultiRegexpStats) {
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

func (p *MultiRegexp) replaceAllUsingOR(text []byte) ([]byte, MultiRegexpStats) {
	// If memoization is enabled, we will use the memoized matches.
	var stats1 MultiRegexpStats
	if p.memoize {
		text, stats1 = p._replaceAllUsingOR(true, text)
	}

	retVal, stats2 := p._replaceAllUsingOR(false, text)
	// In case we had any new matches that were memoized, then we need to recompile
	// the memoized matches regex next time.
	if p.memoize && stats2.numMatches > 0 {
		p.mustRecompileMemoizedMatchesRegex = true
	}

	// Now that we have two stats, let us combine them.
	stats := MultiRegexpStats{}
	stats.numReplacements = stats1.numReplacements + stats2.numReplacements
	stats.numMatches = stats1.numMatches + stats2.numMatches
	stats.numMemoizedMatches = stats1.numMemoizedMatches + stats2.numMemoizedMatches
	stats.regexpMatchDuration = stats1.regexpMatchDuration + stats2.regexpMatchDuration
	stats.numMemoizedEntries = len(p.memoizedMatches)

	return retVal, stats
}

func (p *MultiRegexp) _replaceAllUsingOR(useMemoizedMatchesRegex bool, text []byte) ([]byte, MultiRegexpStats) {
	retVal := make([]byte, 0, len(text))
	stats := MultiRegexpStats{}

	var compiledRegex *regexp.Regexp
	if useMemoizedMatchesRegex && len(p.memoizedMatches) > 0 {
		if p.mustRecompileMemoizedMatchesRegex {
			keys := make([]string, 0, len(p.memoizedMatches))
			for k := range p.memoizedMatches {
				keys = append(keys, k)
			}
			p.compiledMemoizedMatchesRegex = regexp.MustCompile(strings.Join(keys, "|"))
			p.mustRecompileMemoizedMatchesRegex = false
		}
		compiledRegex = p.compiledMemoizedMatchesRegex
	} else {
		compiledRegex = p.compiledRegex
	}

	startTime := time.Now()
	matches := compiledRegex.FindAllSubmatchIndex(text, -1)
	stats.regexpMatchDuration = time.Since(startTime)

	stats.numMatches = len(matches)

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
				stats.numMemoizedMatches++
				stats.numReplacements++
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
				stats.numReplacements++
				break
			}
		}
	}
	retVal = append(retVal, text[lastIndex:]...)
	return retVal, stats
}

func (p *MultiRegexp) replaceAllUsingSubmatch(text []byte) ([]byte, MultiRegexpStats) {
	retVal := make([]byte, 0, len(text))
	stats := MultiRegexpStats{}

	startTime := time.Now()
	matches := p.compiledRegex.FindAllSubmatchIndex(text, -1)
	stats.regexpMatchDuration = time.Since(startTime)

	stats.numMatches = len(matches)

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
			stats.numReplacements++
		}
	}
	retVal = append(retVal, text[lastIndex:]...)
	return retVal, stats
}

func (p *MultiRegexp) replaceAllUsingBruteForce(text []byte) ([]byte, MultiRegexpStats) {
	var retVal []byte
	stats := MultiRegexpStats{}

	for i := range p.compiledRegexList {
		retVal = make([]byte, 0, len(text))

		startTime := time.Now()
		matches := p.compiledRegexList[i].FindAllSubmatchIndex(text, -1)
		stats.regexpMatchDuration += time.Since(startTime)

		stats.numMatches += len(matches)
		replacement := []byte(p.toStrings[i])

		lastIndex := 0

		for _, match := range matches {
			if lastIndex < match[0] {
				retVal = append(retVal, text[lastIndex:match[0]]...)
			}
			retVal = append(retVal, replacement...)
			lastIndex = match[1]
			stats.numReplacements++
		}
		retVal = append(retVal, text[lastIndex:]...)

		text = retVal
	}

	return text, stats
}
