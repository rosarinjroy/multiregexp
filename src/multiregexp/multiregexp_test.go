package parallelreplace

import (
	"fmt"
	"math"
	"os"
	"slices"
	"testing"
)

const NUM_REGEXEX_TO_BENCHMARK = 1000

func TestMultiRegexp(t *testing.T) {
	replacer := NewMultiRegexp()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")

	text := []byte("foo bar baz")
	result := replacer.ReplaceAll(text)
	if string(result) != "bar bar qux" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
}

func TestMultiRegexpWithoutMatch(t *testing.T) {
	replacer := NewMultiRegexp()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")

	text := []byte("bar bar bar")
	result := replacer.ReplaceAll(text)
	if string(result) != "bar bar bar" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
}

func BenchmarkMultiRegexpNoMatch(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXEX_TO_BENCHMARK))))
	replacer := NewMultiRegexp()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	for i := 0; i < NUM_REGEXEX_TO_BENCHMARK; i++ {
		from := fmt.Sprintf(fromFmtStr, i)
		to := fmt.Sprintf(toFmtStr, i+1)
		replacer.AddReplacement(from, to)
	}
	inputText, err := os.ReadFile("oliver_twist.txt")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		outputText := replacer.ReplaceAll(inputText)
		// os.WriteFile(fmt.Sprintf("oliver_twist_%d.txt", i), outputText, 0644)
		if !slices.Equal(outputText, inputText) {
			b.Fatalf("Expected input and output to be the same")
		}
	}
}

func BenchmarkMultiRegexpUsingOR(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXEX_TO_BENCHMARK))))
	replacer := NewMultiRegexp()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := 0; i < NUM_REGEXEX_TO_BENCHMARK; i++ {
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

	b.ResetTimer()
	for range b.N {
		outputText := replacer.ReplaceAll(inputText)
		expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")
		if !slices.Equal(expected, outputText) {
			b.Fatalf("Expected input and output to be the same")
		}
	}
}

func BenchmarkMultiRegexpUsingORWithMemoization(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXEX_TO_BENCHMARK))))
	replacer := NewMultiRegexp()
	replacer.Memoize(true)
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := 0; i < NUM_REGEXEX_TO_BENCHMARK; i++ {
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

	b.ResetTimer()
	for range b.N {
		outputText := replacer.ReplaceAll(inputText)
		expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")
		if !slices.Equal(expected, outputText) {
			b.Fatalf("Expected input and output to be the same")
		}
	}
}

func BenchmarkMultiRegexpUsingSubmatch(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXEX_TO_BENCHMARK))))
	replacer := NewMultiRegexpUsingSubmatch()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := 0; i < NUM_REGEXEX_TO_BENCHMARK; i++ {
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

	b.ResetTimer()
	for range b.N {
		outputText := replacer.ReplaceAll(inputText)
		expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")
		if !slices.Equal(expected, outputText) {
			b.Fatalf("Expected input and output to be the same")
		}
	}
}

func BenchmarkParallelReplacerUsingBruteForce(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(NUM_REGEXEX_TO_BENCHMARK))))
	replacer := NewMultiRegexpUsingBruteForce()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := 0; i < NUM_REGEXEX_TO_BENCHMARK; i++ {
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

	b.ResetTimer()
	for range b.N {
		outputText := replacer.ReplaceAll(inputText)
		expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")
		if !slices.Equal(expected, outputText) {
			b.Fatalf("Expected input and output to be the same")
		}
	}
}
