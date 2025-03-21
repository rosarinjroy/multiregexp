# Problem Statement
I had a need to find and replace a set of regular expression with their respective replacements. These replacements are specified in the form of rules. For e.g.

```
    Rule 0: Replace "regexp1" with "STRING1"
    Rule 1: Replace "regexp2" with "STRING2"
    Rule 2: Replace "regexp3" with "STRING3"
```
If I do the replacement one replacement at a time, I will processing the input string as many times as the number of the rules. This will be inefficient if the input size is vary large.

One option is to form a single regular expression with OR ("|") between each one of them. So we will be matching "regexp1|regexp2|regexp3". Though this method will be faster than multi-pass matching, this method will not work. When the search regular expression matches, we will not be able to tell which one of the input expressions ("regexp1", "regexp2" or "regexp3") matched. Hence we will not be able to identify the correct replacement.

Another option is to form a single regular expression, but make each one of the input regexp as its own submatch. Instead of using "|" to concatenate, we can use ")|(" to concatenate. So we will be matching "(regexp1)|(regexp2)|(regexp3)". In PCRE world, we call the inner matches as a capture groups, where as in Golang world we call them as submatch. They mean the same.

Golang's `regexp` package has a nice function called `FindAllSubmatchIndex`. This function returns `[][]int`. The outer slice has as many elements as there are matches. The inner array represents the whole match and submatches. The elements in the inner array are to be interpreted as below:

- The first pair of integers always represent the full match. The first element of the pair represents the starting position and the second element of the pair represents the length of the match.
- Each subsequent pair represents a subgroup. They are to be interpreted as starting position and length of the match.

So by iterating over the submatches, we will know which replacement should be used. If the submatch is at location 2N of the inner array, we know that we should be replacing that with the replacement string from Rule N-1.

`regexp` package is based on the RE2 library. A nice property of the RE2 library is that it guarantees linear time to match any regular expression (once it is compiled). This property guarantees safety.

# Limitations

- This library does not check if the input expression itself has any submatches yet.
