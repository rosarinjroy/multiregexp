package parallelreplace

import (
	"fmt"
	"regexp"
	"strings"
)

type ParallelReplacer struct {
	fromRegexes []string
	toStrings   []string
}

func NewParallelReplacer() *ParallelReplacer {
	return &ParallelReplacer{}
}

// Adds a new replacement to the list of replacements.
func (p *ParallelReplacer) AddReplacement(old string, new string) {
	p.fromRegexes = append(p.fromRegexes, old)
	p.toStrings = append(p.toStrings, new)
}

func (p *ParallelReplacer) PrintStats() {
	fmt.Println("FromRegexes:", len(p.fromRegexes))
	fmt.Println("ToStrings:", len(p.toStrings))
}

func (p *ParallelReplacer) ReplaceAll(text []byte) []byte {
	// Let us form a regex from the fromRegexes
	regex := "(" + strings.Join(p.fromRegexes, ")|(") + ")"
	// fmt.Println("Regex:",regex)
	re := regexp.MustCompile(regex)
	retVal := make([]byte, 0, len(text))
	matches := re.FindAllSubmatchIndex(text, -1)
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
