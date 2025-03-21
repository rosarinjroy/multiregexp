*** WIP ***

# Problem Statement
I had a need to find and replace a set of regular expression with their respective replacements. These replacements are specified in the form of rules. For e.g.

```
    Rule 0: Replace "regexp1" with "STRING1"
    Rule 1: Replace "regexp2" with "STRING2"
    Rule 2: Replace "regexp3" with "STRING3"
```

The objective is to design the most efficient solution.

# Approach 1 - Multi-pass Replace

If we process the rules one at a time, we will be going over the input string as many times as the number of the rules. This will be inefficient if the number of rules is vary large. I didn't spend time to implement this solution.

# Approach 2 - Single combined regex without submatches
Another option is to form a combined regular expression with OR ("|") between each one of the search patterns. So we will be matching "regexp1|regexp2|regexp3". This method will be faster than multi-pass matching. But this method will require us to identify which one of the rules actually matched, after we receive a combined regex match. This will require us to do a linear search in the list of rules after every combined regex match.

If the number of actual matched strings will be finite, then we can speed up this algorithm by memoization.

# Approach 3 - Single combined regex with submatches

The last option we will explore is to form a combined regular expression, but make each one of the input regexp as its own submatch. Instead of using "|" to concatenate, we can use ")|(" to concatenate. So we will be matching "(regexp1)|(regexp2)|(regexp3)". In PCRE world, we call the inner matches as capture groups, where as in Golang world we call them as submatches. They mean the same.

Golang's `regexp` package has a nice function called [`FindAllSubmatchIndex`](https://pkg.go.dev/regexp#Regexp.FindAllSubmatchIndex). This function returns `[][]int`. The outer slice has as many elements as there are matches. The inner array represents the whole match and submatches. The elements in the inner array are to be interpreted as below:

- The first pair of integers always represent the full match. The first element of the pair represents the starting position and the second element of the pair represents the length of the match.
- Each subsequent pair represents a subgroup. They are to be interpreted as starting position and length of the match.

So by iterating over the submatches, we will know which replacement should be used. If the submatch is at location 2N of the inner array, we know that we should be replacing that with the replacement string from Rule N-1.

`regexp` package is based on the RE2 library. A nice property of the RE2 library is that it guarantees linear time to match any regular expression (once it is compiled). This property guarantees safety.

# Benchmark results

```
$ go test -benchmem -run='^$' -bench '^BenchmarkParallelReplacerWithMatchV2$' rosarinjroy.github.com/parallelreplace/src/parallelreplace -v
goos: darwin
goarch: arm64
pkg: rosarinjroy.github.com/parallelreplace/src/parallelreplace
BenchmarkParallelReplacerWithMatchV2
BenchmarkParallelReplacerWithMatchV2-8   	  422474	      2884 ns/op	    3079 B/op	      28 allocs/op
PASS
ok  	rosarinjroy.github.com/parallelreplace/src/parallelreplace	1.658s

$ go test -benchmem -run='^$' -benchtime=5s -bench '^BenchmarkParallelReplacerWithMatch$' rosarinjroy.github.com/parallelreplace/src/parallelreplace -v
goos: darwin
goarch: arm64
pkg: rosarinjroy.github.com/parallelreplace/src/parallelreplace
BenchmarkParallelReplacerWithMatch
BenchmarkParallelReplacerWithMatch-8   	     126	  59018498 ns/op	   29529 B/op	      10 allocs/op
PASS
ok  	rosarinjroy.github.com/parallelreplace/src/parallelreplace	12.270s
```

# Limitations

- This library does not check if the input expression itself has any submatches yet.
- If the input has any 
