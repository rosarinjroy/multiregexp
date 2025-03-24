*** WIP ***

# Problem Statement
I had a need to find and replace a set of regular expression with their respective replacement strings. These replacements are specified in the form of rules. For e.g.

```
    Rule 0: Replace "regexp1" with "STRING1"
    Rule 1: Replace "regexp2" with "STRING2"
    Rule 2: Replace "regexp3" with "STRING3"
```

The objective is to design the most efficient solution. I could come up with three different approaches that are explained below. This library has code, tests and benchmarks for all the three approaches.

# Approach 1 - Single combined regex without submatches
First option is to form a combined regular expression with OR ("|") between each one of the search patterns. So we will be matching "regexp1|regexp2|regexp3". This method will be faster than multi-pass matching. But this method will require us to identify which one of the rules actually matched, after we receive a combined regex match. This will require us to do a linear search in the list of rules after every combined regex match.

If the number of actual matched strings will be finite, then we can speed up this algorithm by memoization.

# Approach 2 - Single combined regex with submatches

Another option is to form a combined regular expression, but make each one of the input regexp as its own submatch. Instead of using "|" to concatenate, we can use ")|(" to concatenate. So we will be matching "(regexp1)|(regexp2)|(regexp3)". In PCRE world, we call the inner matches as capture groups, where as in Golang world we call them as submatches. They mean the same.

# How to find the matching region

Golang's `regexp` package has a nice function called [`FindAllSubmatchIndex`](https://pkg.go.dev/regexp#Regexp.FindAllSubmatchIndex). This function returns `[][]int`. The outer slice has as many elements as there are matches. The inner slice represents the whole match and submatches. The elements in the inner array are to be interpreted as below:

- The first pair of integers always represent the full match. The first element of the pair represents the starting position (inclusive) and the second element of the pair represents the ending position (exclusive) of the match.
- Each subsequent pair represents a subgroup. They are to be interpreted as starting position and ending position of the match.

So by iterating over the submatches, we will know which replacement should be used. If the submatch is at location 2N of the inner array, we know that we should be replacing that with the replacement string from Rule N-1. When there are no submatches (as is the case for approach 1), we should only be looking at the elements at positions 0 and 1 that represent the full match.

`regexp` package is based on the RE2 library. A nice property of the RE2 library is that it guarantees linear time to match any regular expression (once it is compiled). This property guarantees safety.

# Approach 3 - Brute force aka multi-pass replace

For the sake of comparison, I also decided to include brute force method that takes each input rule and executes that rule in all input text. This approach will go over the input string as many times as the number of the rules. This will be inefficient if the number of rules is vary large.

# Benchmark results

```
$ go test -benchmem -run='^$' -benchtime=5s -bench '^Benchmark' ./src/multiregexp/ -v
goos: darwin
goarch: arm64
pkg: rosarinjroy.github.com/multiregexp/src/multiregexp
BenchmarkMultiRegexpNoMatch
BenchmarkMultiRegexpNoMatch-8                  	    8610	    706499 ns/op	  958989 B/op	       4 allocs/op
BenchmarkMultiRegexpUsingOR
BenchmarkMultiRegexpUsingOR-8                  	     256	  23278466 ns/op	 1930135 B/op	     107 allocs/op
BenchmarkMultiRegexpUsingORWithMemoization
BenchmarkMultiRegexpUsingORWithMemoization-8   	     259	  23099182 ns/op	 1925550 B/op	      96 allocs/op
BenchmarkMultiRegexpUsingSubmatch
BenchmarkMultiRegexpUsingSubmatch-8            	       1	363420208750 ns/op	37521832 B/op	    8161 allocs/op
BenchmarkParallelReplacerUsingBruteForce
BenchmarkParallelReplacerUsingBruteForce-8     	      14	 388919893 ns/op	961567181 B/op	    3567 allocs/op
PASS
ok  	rosarinjroy.github.com/multiregexp/src/multiregexp	396.467s
```

# Limitations

- This library does not check if the input expression itself has any submatches yet.
- If the input has any
