# Problem Statement
I had a need to find a set of regular expressions and replace them with some other strings. Each regular expression had its corresponding replacement. For e.g.

```
    # Repacements
    0: Replace "regexp1" with "STRING1"
    1: Replace "regexp2" with "STRING2"
    2: Replace "regexp3" with "STRING3"
```
If we do the replacement one replacement at a time, we will passing through the input string as many times as the number of the repacements. This will be inefficient if the input size is vary large.

One option is to form a single regular expression with OR ("|") between each one of them. So we will be matching "regexp1|regexp2|regexp3". This will not work. When the full regular expression matches, we will not be able to tell which one of the input expressions ("regexp1", "regexp2" or "regexp3" matched).

Another option is to form a single regular expression, but make each one of the input regexp as its own submatch. I.e. instead of using "|" to concatenate, ")|(" to concatenate. So we will be matching "(regexp1)|(regexp2)|(regexp3)". Golang's `regexp` package has a nice function called `FindAllSubmatchIndex`. This function returns `[][]int`. The inner array represents the matches. The elements in the inner array are to be interpreted as below:

- Each pair of integers represent a subgroup. The first integer is the location (byte offset) and the second integer is the length of the match starting at the byte offset.
- The first pair of integers always represent the full match.
- The second pair onwards represent each of the submatches. They are the matches between each pairs of "(" and ")".

So by iterating over the submatches, we will know which replacement should be used. If the submatch is at location 2N of the inner array, we know that we should be replacing that with the replacement string at N-1 in the replacements.

`regexp` package is based on the RE2 library. A nice property of the RE2 library is that it guarantees linear time to match any regular expression (once it is compiled).

# Limitations

- This library does not check if the input expression itself has any submatches yet.
