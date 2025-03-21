package parallelreplace

import (
	"fmt"
	"regexp"
	"strings"
)

type ParallelReplacer struct {
	fromRegexes       []string
	toStrings         []string
	compiledRegex     *regexp.Regexp
	compiledRegexV2   *regexp.Regexp
	compiledRegexList []*regexp.Regexp
}

func NewParallelReplacer() *ParallelReplacer {
	return &ParallelReplacer{}
}

// Adds a new replacement to the list of replacements.
func (p *ParallelReplacer) AddReplacement(old string, new string) {
	p.fromRegexes = append(p.fromRegexes, old)
	p.toStrings = append(p.toStrings, new)
	p.compiledRegex = nil
	p.compiledRegexV2 = nil
	p.compiledRegexList = nil
}

func (p *ParallelReplacer) Compile() {
	if p.compiledRegex == nil {
		p.compiledRegex = regexp.MustCompile("(" + strings.Join(p.fromRegexes, ")|(") + ")")
	}
}

func (p *ParallelReplacer) CompileV2() {
	if p.compiledRegexV2 == nil {
		p.compiledRegexV2 = regexp.MustCompile(strings.Join(p.fromRegexes, "|"))
	}
	if p.compiledRegexList == nil {
		for _, old := range p.fromRegexes {
			p.compiledRegexList = append(p.compiledRegexList, regexp.MustCompile(old))
		}
	}
}
func (p *ParallelReplacer) PrintStats() {
	fmt.Println("FromRegexes:", len(p.fromRegexes))
	fmt.Println("ToStrings:", len(p.toStrings))
}

func (p *ParallelReplacer) ReplaceAll(text []byte) []byte {
	p.Compile()
	retVal := make([]byte, 0, len(text))
	matches := p.compiledRegex.FindAllSubmatchIndex(text, -1)
	// fmt.Println(re.NumSubexp())
	lastIndex := 0
	for _, match := range matches {
		// fmt.Println(match)
		for i := 2; i < len(match); i += 2 {
			if match[i] != -1 {
				if lastIndex < match[i] {
					retVal = append(retVal, text[lastIndex:match[i]]...)
				}
				replacement := []byte(p.toStrings[(i-2)/2])
				retVal = append(retVal, replacement...)
				lastIndex = match[i] + len(replacement)
				break
			}
		}
	}
	retVal = append(retVal, text[lastIndex:]...)
	return retVal
}

func (p *ParallelReplacer) ReplaceAllV2(text []byte) []byte {
	p.CompileV2()
	retVal := make([]byte, 0, len(text))
	matches := p.compiledRegexV2.FindAllSubmatchIndex(text, -1)
	// fmt.Println(re.NumSubexp())
	lastIndex := 0
	for _, match := range matches {
		// fmt.Println(match)
		if lastIndex < match[0] {
			retVal = append(retVal, text[lastIndex:match[0]]...)
		}
		for j := range p.compiledRegexList {
			if p.compiledRegexList[j].Match(text[match[0]:match[1]]) {
				replacement := []byte(p.toStrings[j])
				retVal = append(retVal, replacement...)
				lastIndex = match[0] + len(replacement)
				break
			}
		}
	}
	retVal = append(retVal, text[lastIndex:]...)
	return retVal
}
