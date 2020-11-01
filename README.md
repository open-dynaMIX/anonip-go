# anonip-go

[![Build Status](https://github.com/open-dynaMIX/anonip-go/workflows/Tests/badge.svg)](https://github.com/open-dynaMIX/anonip-go/actions?query=workflow%3ATests)
[![Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen.svg)](https://github.com/open-dynaMIX/anonip-go/blob/master/anonip_test.go#L22)
[![goreport](https://goreportcard.com/badge/github.com/open-dynaMIX/anonip-go)](https://goreportcard.com/report/github.com/open-dynaMIX/anonip-go)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)

**WIP**

[Anonip](https://github.com/DigitaleGesellschaft/Anonip) rewritten in go.

This is WIP and mainly serves an educational purpose at this time.

## Usage

```
Usage: anonip [--ipv4mask INTEGER] [--ipv6mask INTEGER] [--increment INTEGER] [--output FILE] [--input FILE] [--columns INTEGER [INTEGER ...]] [--delimiter STRING] [--replace STRING] [--regex STRING [STRING ...]] [--skip-private] [--version]

Options:
  --ipv4mask INTEGER, -4 INTEGER
                         truncate the last n bits [default: 12]
  --ipv6mask INTEGER, -6 INTEGER
                         truncate the last n bits [default: 84]
  --increment INTEGER, -i INTEGER
                         increment the IP address by n [default: 0]
  --output FILE, -o FILE
                         file or FIFO to write to [default: stdout]
  --input FILE           file or FIFO to read from [default: stdin]
  --columns INTEGER [INTEGER ...], -c INTEGER [INTEGER ...]
                         assume IP address is in column n (1-based indexed) [default: 0]
  --delimiter STRING, -l STRING
                         log delimiter [default:  ]
  --replace STRING, -r STRING
                         replacement string in case address parsing fails (Example: 0.0.0.0)
  --regex STRING [STRING ...]
                         regex
  --skip-private, -p     do not mask addresses in private ranges. See IANA Special-Purpose Address Registry [default: false]
  --version, -v          show program's version number and exit [default: false]
  --help, -h             display this help and exit
```