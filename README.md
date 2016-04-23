# snip
snip, cut, trim, chop: a lovechild of grep and sed.

## Installation

```
$ GOPATH=$PWD go install github.com/danmarg/snip
```

## Usage

`snip` can do some of what `grep`, `sed`, and `cut` can do, but with your familiar standard `re2` regular expression language.

```
COMMANDS:
    match, m    [pattern] [file]? regular expression match
    replace, s  [pattern] [pattern] [file]? regular expression replace
    split, c    [pattern] [file]? split input lines

GLOBAL OPTIONS:
   --insensitive, -i    case insensitive
   --multiline, -m      multiline
   --dotall, -s         let . match \n
   --ungreedy, -U       swap meaning of x* and x*?, x+ and x+?
```

A few lame examples:

```
$ snip -h | snip m exp                                                                                             
    match, m    [pattern] [file]? regular expression match
    replace, s  [pattern] [pattern] [file]? regular expression replace
$ snip -h | snip s expression exposition | snip m exp                                                              
    match, m    [pattern] [file]? regular exposition match
    replace, s  [pattern] [pattern] [file]? regular exposition replace
$ snip -h | snip m , | snip c , -f 2                                                                               
 cut
 m      [pattern] [file]? regular expression match
 s      [pattern] [pattern] [file]? regular expression replace
 c      [pattern] [file]? split input lines
 -i     case insensitive
 -m     multiline
 -s             let . match \n
 -U     swap meaning of x* and x*?
 -h             show help
 -v     print the version
 ```
