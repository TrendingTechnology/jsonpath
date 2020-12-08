# JSONPath

[![Build Status](https://travis-ci.com/AsaiYusuke/jsonpath.svg?branch=main)](https://travis-ci.com/AsaiYusuke/jsonpath)
[![Go Report Card](https://goreportcard.com/badge/github.com/AsaiYusuke/jsonpath)](https://goreportcard.com/report/github.com/AsaiYusuke/jsonpath)
[![Coverage Status](https://coveralls.io/repos/github/AsaiYusuke/jsonpath/badge.svg?branch=main)](https://coveralls.io/github/AsaiYusuke/jsonpath?branch=main)
[![Go doc](https://godoc.org/github.com/AsaiYusuke/jsonpath?status.svg)](https://pkg.go.dev/github.com/AsaiYusuke/jsonpath)

This is [Go](https://golang.org/) package providing the features that retrieves a part of the JSON objects according to the statement written in the JSONPath syntax.

The core syntaxes of the JSONPath on which this package is based:  [JSONPath - XPath for JSON](https://goessner.net/articles/JsonPath/).

#### Note:
The unstated syntaxes found in "JSONPath - XPath for JSON" are implemented with reference to the test cases written in [cburmer's json-path-comparison](https://github.com/cburgmer/json-path-comparison).
Please check [my compare result](https://asaiyusuke.github.io/jsonpath/cburgmer-json-path-comparison/docs/index.html) to know which responses are adapted.
Unfortunately, the proposals that is also discussing in "json-path-comparison" were not finalized at the start of development and were not adopted outright.

## Getting started

```bash
go get github.com/AsaiYusuke/jsonpath
```

### Simple example:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/AsaiYusuke/jsonpath"
)

func main() {
	jsonPath, srcJSON := `$.key`, `{"key":"value"}`
	var src interface{}
	json.Unmarshal([]byte(srcJSON), &src)
	output, _ := jsonpath.Retrieve(jsonPath, src)
	outputJSON, _ := json.Marshal(output)
	fmt.Println(string(outputJSON))
	// Output:
	// ["value"]
}
```

## Basic design

- [PEG](https://github.com/pointlander/peg) separated the JSONPath syntax analyzer from functionality itself to simplify the source.
- The error specification allows package users to handle errors appropriately.
- Adopted more of the consensus behavior from the [cburmer's json-path-comparison](https://github.com/cburgmer/json-path-comparison).
  Adapted my own behavior to the other part of the such consensus behavior that found difficult to use.
- Equipped with numerous unit tests and tried to eliminate the bugs that return strange result.

## How to use

### * Retrieve one-time, or successively

The `Retrieve` function returns a retrieved JSON object by a one-time sequential operation (analyzing syntax and retrieving objects) using the given JSONPath and the source JSON object :

```go
output, err := jsonpath.Retrieve(jsonPath, src)
```

The `Parse` function returns a *parser-function* that completed to analyze the JSONPath syntax.
By using this returned *parser-function* it can be performed successively a retrieve with the same JSONPath syntax :

```go
jsonPath, err := jsonpath.Parse(jsonPath)
output1, err1 := jsonPath(src1)
output2, err2 := jsonPath(src2)
:
```

### * Error handling

If there is a problem with the execution of the `Retrieve`, `Prepare` or prepared *parser-functions*, an error type is returned.
These error types define the corresponding symptom, as listed below:

#### Syntax analyze errors from `Retrieve`, `Parse`

| Error type             | Message format                                     | Symptom                                                                                                       |
|------------------------|----------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| `ErrorInvalidSyntax`   | `invalid syntax (position=%d, reason=%s, near=%s)` | The invalid syntax found in the JSONPath. The *reason* including in this message will tell you more about it. |
| `ErrorInvalidArgument` | `invalid argument (argument=%s, error=%s)`         | The argument specified in the JSONPath was treated as the invalid error in Go syntax.                         |
| `ErrorNotSupported`    | `not supported (feature=%s, path=%s)`              | The unsupported syntaxes specified in the JSONPath.                                                           |

#### Runtime errors from `Retrieve`, *`parser-functions`*

| Error type             | Message format                                    | Symptom                                                                        |
|------------------------|---------------------------------------------------|--------------------------------------------------------------------------------|
| `ErrorMemberNotExist`  | `member did not exist (path=%s)`                  | The object member specified in the JSONPath did not exist in the JSON object.  |
| `ErrorIndexOutOfRange` | `index out of range (path=%s)`                    | The array indexes specified in the JSONPath were out of range.                 |
| `ErrorTypeUnmatched`   | `type unmatched (expected=%s, found=%s, path=%s)` | The node type specified in the JSONPath did not exist in the JSON object.      |
| `ErrorNoneMatched`     | `none matched (path=%s)`                          | The retrieving child paths specified in the JSONPath resulted in empty output. |

The type checking is convenient to recognize which error happened.

```go
:
_,err := jsonpath.Retrieve(jsonPath, srcJSON)
if err != nil {
  switch err.(type) {
  case jsonpath.ErrorIndexOutOfRange:
    fmt.printf(`retry with other srcJSON: %v`, err)
    continue
  case jsonpath.ErrorInvalidArgumentFormat:
    return nil, fmt.errorf(`specified invalid argument: %v`, err)
  }
  :
}
```

## Differences

Some behaviors that differ from the consensus exists in this package.
For the entire comparisons, please check [this result](https://asaiyusuke.github.io/jsonpath/cburgmer-json-path-comparison/docs/index.html) to see which responses are different.

### Character types

The character types that can not be used for the identifiers in the dot notation are as follows :

```text
. [ ) = ! > < \t \r \n *SPACE*
```

You have to encode these characters when you enter them :

```text
JSONPath : $.abc\.def
srcJSON  : {"abc.def":1}
Output   : 1
```

### Asterisk in qualifier

The asterisk in qualifier can be specified mixed with other subscript syntaxes.

```text
JSONPath : $[0,1:3,*]
srcJSON  : [0,1,2,3,4,5]
Output   : [0,1,2,0,1,2,3,4,5]
```

### Regular expression

The regular expression syntax works as a regular expression in Go lang.
In particular, you can use "(?i)" to specify the regular expression as the ignore case option.

```text
JSONPath : $[?(@.a=~/(?i)CASE/)]
srcJSON  : ["Case","Hello"]
Output   : ["Case"]
```

## Benchmarks

<details>
<summary>Show results</summary>

```text
goos: windows
goarch: amd64
pkg: github.com/AsaiYusuke/jsonpath
BenchmarkRetrieve_dotNotation-4                             	   11354	     98725 ns/op	  533608 B/op	     146 allocs/op
BenchmarkRetrieve_bracketNotation-4                         	   10000	    119519 ns/op	  533623 B/op	     147 allocs/op
BenchmarkRetrieve_asterisk_identifier_dotNotation-4         	   10000	    105806 ns/op	  533651 B/op	     148 allocs/op
BenchmarkRetrieve_asterisk_identifier_bracketNotation-4     	   10000	    118763 ns/op	  533592 B/op	     146 allocs/op
BenchmarkRetrieve_multi_identifier-4                        	    8020	    129579 ns/op	  533902 B/op	     157 allocs/op
BenchmarkRetrieve_qualifier_index-4                         	   10000	    114525 ns/op	  533598 B/op	     148 allocs/op
BenchmarkRetrieve_qualifier_slice-4                         	    9230	    120536 ns/op	  533805 B/op	     152 allocs/op
BenchmarkRetrieve_qualifier_asterisk-4                      	   10000	    108871 ns/op	  533561 B/op	     146 allocs/op
BenchmarkRetrieve_qualifier_union-4                         	   10000	    123751 ns/op	  533890 B/op	     159 allocs/op
BenchmarkRetrieve_filter_logicalOR-4                        	    9246	    134633 ns/op	  534613 B/op	     161 allocs/op
BenchmarkRetrieve_filter_logicalAND-4                       	   10000	    132457 ns/op	  534869 B/op	     163 allocs/op
BenchmarkRetrieve_filter_nodeFilter-4                       	    9250	    125146 ns/op	  534417 B/op	     161 allocs/op
BenchmarkRetrieve_filter_logicalNOT-4                       	    9256	    133314 ns/op	  534485 B/op	     164 allocs/op
BenchmarkRetrieve_filter_compareEQ-4                        	    8593	    143504 ns/op	  535122 B/op	     171 allocs/op
BenchmarkRetrieve_filter_compareNE-4                        	    9252	    141937 ns/op	  534932 B/op	     171 allocs/op
BenchmarkRetrieve_filter_compareGE-4                        	    8594	    143957 ns/op	  535120 B/op	     171 allocs/op
BenchmarkRetrieve_filter_compareGT-4                        	    8592	    142375 ns/op	  535013 B/op	     170 allocs/op
BenchmarkRetrieve_filter_compareLE-4                        	    8593	    143808 ns/op	  535106 B/op	     170 allocs/op
BenchmarkRetrieve_filter_compareLT-4                        	    9255	    142373 ns/op	  535007 B/op	     169 allocs/op
BenchmarkRetrieve_filter_regex-4                            	    7519	    147486 ns/op	  543410 B/op	     183 allocs/op
BenchmarkParserFunc_dotNotation-4                           	 9400027	       126 ns/op	      48 B/op	       2 allocs/op
BenchmarkParserFunc_bracketNotation-4                       	 9473965	       126 ns/op	      48 B/op	       2 allocs/op
BenchmarkParserFunc_asterisk_identifier_dotNotation-4       	 4282303	       277 ns/op	      96 B/op	       4 allocs/op
BenchmarkParserFunc_asterisk_identifier_bracketNotation-4   	 4198162	       286 ns/op	      96 B/op	       4 allocs/op
BenchmarkParserFunc_multi_identifier-4                      	 5104863	       238 ns/op	      80 B/op	       3 allocs/op
BenchmarkParserFunc_qualifier_index-4                       	 6780312	       178 ns/op	      64 B/op	       4 allocs/op
BenchmarkParserFunc_qualifier_slice-4                       	 5805200	       207 ns/op	      64 B/op	       4 allocs/op
BenchmarkParserFunc_qualifier_asterisk-4                    	 6585810	       178 ns/op	      64 B/op	       4 allocs/op
BenchmarkParserFunc_qualifier_union-4                       	 3584234	       335 ns/op	     120 B/op	       7 allocs/op
BenchmarkParserFunc_filter_logicalOR-4                      	  999666	      1184 ns/op	     912 B/op	      12 allocs/op
BenchmarkParserFunc_filter_logicalAND-4                     	  858484	      1364 ns/op	    1168 B/op	      14 allocs/op
BenchmarkParserFunc_filter_nodeFilter-4                     	 1594872	       741 ns/op	     608 B/op	       8 allocs/op
BenchmarkParserFunc_filter_logicalNOT-4                     	 1218441	       974 ns/op	     656 B/op	      10 allocs/op
BenchmarkParserFunc_filter_compareEQ-4                      	  925389	      1328 ns/op	    1120 B/op	      12 allocs/op
BenchmarkParserFunc_filter_compareNE-4                      	 1000000	      1217 ns/op	     912 B/op	      11 allocs/op
BenchmarkParserFunc_filter_compareGE-4                      	  940814	      1242 ns/op	    1120 B/op	      12 allocs/op
BenchmarkParserFunc_filter_compareGT-4                      	 1000000	      1242 ns/op	    1120 B/op	      12 allocs/op
BenchmarkParserFunc_filter_compareLE-4                      	  933510	      1232 ns/op	    1120 B/op	      12 allocs/op
BenchmarkParserFunc_filter_compareLT-4                      	  859462	      1247 ns/op	    1120 B/op	      12 allocs/op
BenchmarkParserFunc_filter_regex-4                          	  801913	      1431 ns/op	    1137 B/op	      13 allocs/op
```
</details>

## Project progress

- syntax
  - identifier
    - [x] identifier in dot notations
    - [x] identifier in bracket notations
    - [x] asterisk identifier
    - [x] multiple-identifier in bracket
    - [x] recursive retrieve
  - qualifier
    - [x] index
    - [x] slice
    - [x] asterisk
    - filter
      - [x] logical operation
      - [x] comparator
      - [x] JSONPath retrieve in filter
    - [ ] script
  - [x] Refer to the consensus behaviors
- archtecture
  - [x] PEG syntax analyzing
  - [x] Error handling
- Go language manner
  - [x] retrieve with the object in interface unmarshal
  - [x] retrieve with the json.Number type
- source code
  - [x] Release version
  - Unit tests
    - [x] syntax tests
    - [x] benchmark
    - [x] coverage >80%
  - [ ] Examples
  - [x] CI automation
  - Documentation
    - [x] README
    - [ ] API doc
  - [x] comparison result (local)
- future todo
  - [ ] Syntax expansion
  - [ ] Refer to the something standard
  - Go language affinity
    - [ ] retrieve with the object in struct unmarshal
    - [ ] retrieve with the struct tags
    - [ ] retrieve with the user defined objects
