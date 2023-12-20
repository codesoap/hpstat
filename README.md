hpstat takes lines in the [httpipe
format](https://github.com/codesoap/httpipe) via standard input and
prints them, if they match a given status code filter. If no filter is
given, hpstat will count how often each status code occurs.

# Examples
```console
$ cat /path/to/saved/httpipe | hpstat
Invalid lines: 3
Status code 200: 1013x
Status code 301: 11x
Status code 404: 2817x
Status code 500: 1x

$ # Investigate the 500 response:
$ cat /path/to/saved/httpipe | hpstat 500 | jq .
...

$ # Extract all non-2xx responses:
$ cat /path/to/saved/httpipe | hpstat -v 200:299 > /path/to/new/file
```

# Usage
```console
$ hpstat -h
Usage: hpstat [-v] [FILTER]...

If no argument is given, the amounts of occurrences of each status code
are counted and displayed.

If one or more arguments are given, lines are filtered and printed only
if they match the filter. Arguments may either be single status codes,
like '200' or ranges, like '200:299'.

If the -v flag is given, the filter is inverted. Only lines with status
codes, that don't match the given arguments, will be printed.
```
