package parallelreplace

import (
	"fmt"
	"math"
	"os"
	"slices"
	"testing"
)

func TestParallelReplacer(t *testing.T) {
	replacer := NewParallelReplacer()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")

	text := []byte("foo bar baz")
	result := replacer.ReplaceAll(text)
	if string(result) != "bar bar qux" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
}

func TestParallelReplacerWithoutMatch(t *testing.T) {
	replacer := NewParallelReplacer()
	replacer.AddReplacement("foo", "bar")
	replacer.AddReplacement("baz", "qux")

	text := []byte("bar bar bar")
	result := replacer.ReplaceAll(text)
	if string(result) != "bar bar bar" {
		t.Errorf("Expected 'bar bar qux', got '%s'", string(result))
	}
}

func BenchmarkParallelReplacerNoMatch(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(b.N))))
	replacer := NewParallelReplacer()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	for i := 0; i < b.N; i++ {
		from := fmt.Sprintf(fromFmtStr, i)
		to := fmt.Sprintf(toFmtStr, i+1)
		replacer.AddReplacement(from, to)
	}
	inputText, err := os.ReadFile("oliver_twist.txt")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range 100 {
		outputText := replacer.ReplaceAll(inputText)
		// os.WriteFile(fmt.Sprintf("oliver_twist_%d.txt", i), outputText, 0644)
		if !slices.Equal(outputText, inputText) {
			b.Fatalf("Expected input and output to be the same")
		}
	}
}

func BenchmarkParallelReplacerWithMatch(b *testing.B) {
	numDigits := int(math.Ceil(math.Log10(float64(b.N))))
	replacer := NewParallelReplacer()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := 0; i < b.N; i++ {
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

	outputText := replacer.ReplaceAll(inputText)
	expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")
	if !slices.Equal(expected, outputText) {
		b.Fatalf("Expected input and output to be the same")
	}

}

func BenchmarkParallelReplacerWithMatchV2(b *testing.B) {
	if b.N > 1000000 {
		b.Skipf("Will not run the benchmark for N > 10000, N = %d", b.N)
	}

	numDigits := int(math.Ceil(math.Log10(float64(b.N))))
	replacer := NewParallelReplacer()
	fromFmtStr := fmt.Sprintf("foo%%0%dd", numDigits)
	toFmtStr := fmt.Sprintf("bar%%0%dd", numDigits)
	// fmt.Println("Input:", b.N)
	for i := 0; i < b.N; i++ {
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

	outputText := replacer.ReplaceAllV2(inputText)
	expected, _ := os.ReadFile("oliver_twist_XXXXXXX.golden")
	if !slices.Equal(expected, outputText) {
		b.Fatalf("Expected input and output to be the same")
	}

}
