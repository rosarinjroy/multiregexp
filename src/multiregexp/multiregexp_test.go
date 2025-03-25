package parallelreplace

import (
	"fmt"
	"math"
	"os"
	"slices"
	"testing"
)

const NUM_REGEXPS_TO_BENCHMARK = 1

func TestMultiRegexpUsingOR(t *testing.T) {
	replacer := NewMultiRegexp()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")

	text := []byte("foo bar baz")
	result, _ := replacer.ReplaceAll(text)
	if string(result) != "bar bar qux" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
}

func TestMultiRegexpUsingORWithMemoization(t *testing.T) {
	replacer := NewMultiRegexp()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")
	replacer.Memoize(true)

	text := []byte("foo bar baz")
	result, stats := replacer.ReplaceAll(text)
	if string(result) != "bar bar qux" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
	if stats.numReplacements != 2 {
		t.Errorf("Expected 2 replacements, got %d", stats.numReplacements)
	}
	if stats.numMemoizedMatches != 0 {
		t.Errorf("Expected 1 memoized match, got %d", stats.numMemoizedMatches)
	}

	result, stats = replacer.ReplaceAll(text)
	if string(result) != "bar bar qux" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
	if stats.numReplacements != 2 {
		t.Errorf("Expected 2 replacements, got %d", stats.numReplacements)
	}
	if stats.numMemoizedMatches != 2 {
		t.Errorf("Expected 1 memoized match, got %d", stats.numMemoizedMatches)
	}
}

func TestMultiRegexpUsingORWithoutMatch(t *testing.T) {
	replacer := NewMultiRegexp()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")

	text := []byte("bar bar bar")
	result, _ := replacer.ReplaceAll(text)
	if string(result) != "bar bar bar" {
		t.Errorf("Expected 'bar bar bar', got '%s'", string(result))
	}
}

func TestMultiRegexpUsingSubmatch(t *testing.T) {
	replacer := NewMultiRegexpUsingSubmatch()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")

	text := []byte("foo bar baz")
	result, _ := replacer.ReplaceAll(text)
	if string(result) != "bar bar qux" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
}

func TestMultiRegexpUsingBruteForce(t *testing.T) {
	replacer := NewMultiRegexpUsingBruteForce()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")

	text := []byte("foo bar baz")
	result, _ := replacer.ReplaceAll(text)
	if string(result) != "bar bar qux" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
}
func BenchmarkMultiRegexpNoMatch(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXPS_TO_BENCHMARK))))
	replacer := NewMultiRegexp()
	replacer.AddReplacement("XXXXXXX", "Charles")
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	for i := range NUM_REGEXPS_TO_BENCHMARK {
		from := fmt.Sprintf(fromFmtStr, i)
		to := fmt.Sprintf(toFmtStr, i)
		replacer.AddReplacement(from, to)
	}

	inputText, err := os.ReadFile("oliver_twist.txt")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		outputText, _ := replacer.ReplaceAll(inputText)
		// os.WriteFile(fmt.Sprintf("oliver_twist_%d.txt", i), outputText, 0644)
		if !slices.Equal(outputText, inputText) {
			b.Fatalf("Expected output does not match with actual output")
		}
	}
}

func BenchmarkMultiRegexpUsingOR(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXPS_TO_BENCHMARK))))
	replacer := NewMultiRegexp()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := range NUM_REGEXPS_TO_BENCHMARK {
		from := fmt.Sprintf(fromFmtStr, i)
		to := fmt.Sprintf(toFmtStr, i+1)
		replacer.AddReplacement(from, to)
	}
	replacer.AddReplacement("Charles", "XXXXXXX")
	// replacer.PrintStats()
	inputText, err := os.ReadFile("oliver_twist.txt")
	if err != nil {
		b.Fatal(err)
	}
	expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")

	b.ResetTimer()
	for range b.N {
		outputText, stats := replacer.ReplaceAll(inputText)
		// fmt.Printf("stats: %+v\n", stats)
		if !slices.Equal(expected, outputText) {
			b.Fatalf("Expected output does not match with actual output")
		}
		if stats.numMemoizedMatches > 0 {
			b.Fatalf("Expected zero memoized matches, but got numMemoizedMatches = %d", stats.numMemoizedMatches)
		}
	}
}

func BenchmarkMultiRegexpUsingORWithMemoization(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXPS_TO_BENCHMARK))))
	replacer := NewMultiRegexp()
	replacer.Memoize(true)
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := range NUM_REGEXPS_TO_BENCHMARK {
		from := fmt.Sprintf(fromFmtStr, i)
		to := fmt.Sprintf(toFmtStr, i+1)
		replacer.AddReplacement(from, to)
	}
	replacer.AddReplacement("Charles", "XXXXXXX")
	// replacer.PrintStats()
	inputText, err := os.ReadFile("oliver_twist.txt")
	if err != nil {
		b.Fatal(err)
	}
	expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")

	b.ResetTimer()
	for i := range b.N {
		outputText, stats := replacer.ReplaceAll(inputText)
		if !slices.Equal(expected, outputText) {
			b.Fatalf("Expected output does not match with actual output: len(expected) = %d, len(outputText) = %d", len(expected), len(outputText))
		}
		if i == 0 || i == b.N-1 {
			fmt.Printf("i = %d, stats: %+v\n", i, stats)
		}
		if i > 0 {
			if stats.numReplacements != stats.numMemoizedMatches {
				b.Fatalf("numReplacements = %d, numMemoizedMatches = %d", stats.numReplacements, stats.numMemoizedMatches)
			}
		}
	}
}

func BenchmarkMultiRegexpUsingSubmatch(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXPS_TO_BENCHMARK))))
	replacer := NewMultiRegexpUsingSubmatch()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := range NUM_REGEXPS_TO_BENCHMARK {
		from := fmt.Sprintf(fromFmtStr, i)
		to := fmt.Sprintf(toFmtStr, i+1)
		replacer.AddReplacement(from, to)
	}
	replacer.AddReplacement("Charles", "XXXXXXX")
	// replacer.PrintStats()
	inputText, err := os.ReadFile("oliver_twist.txt")
	if err != nil {
		b.Fatal(err)
	}
	expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")

	b.ResetTimer()
	for range b.N {
		outputText, _ := replacer.ReplaceAll(inputText)
		if !slices.Equal(expected, outputText) {
			b.Fatalf("Expected output does not match with actual output")
		}
	}
}

func BenchmarkParallelReplacerUsingBruteForce(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXPS_TO_BENCHMARK))))
	replacer := NewMultiRegexpUsingBruteForce()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := range NUM_REGEXPS_TO_BENCHMARK {
		from := fmt.Sprintf(fromFmtStr, i)
		to := fmt.Sprintf(toFmtStr, i+1)
		replacer.AddReplacement(from, to)
	}
	replacer.AddReplacement("Charles", "XXXXXXX")
	// replacer.PrintStats()
	inputText, err := os.ReadFile("oliver_twist.txt")
	if err != nil {
		b.Fatal(err)
	}
	expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")

	b.ResetTimer()
	for range b.N {
		outputText, _ := replacer.ReplaceAll(inputText)

		if !slices.Equal(expected, outputText) {
			b.Fatalf("Expected output does not match with actual output")
		}
	}
}
