findjson
---------

Find json documents contained within a stream. Useful for parsing json documents from log files.

Consider this piece of text:

`error writing file {"func": "OpenFile", "reason": "file not found", "code": 500}`

There is a json object in the text above. This library includes a thread-safe parser that finds the (start, end) index(es) of valid json objects. It also writes these JSON objects into an output stream. The output of this program with the above input will be

`{"func": "OpenFile", "reason": "file not found", "code": 500}`

Download
---------

Download the CLI that filters out JSON objects out of any text stream

```bash
# get the binary
$ wget https://github.com/wlan0/findjson/releases/download/v0.1/findjson 
# change permissions
$ chmod +x ./findjson
# Test it with an example text stream (example has a 180MB JSON file)
$ curl -sSL https://github.com/zemirco/sf-city-lots-json/blob/master/citylots.json?raw=true | ./findjson
```


Usage
--------

Just pass a `io.Reader` and `io.Writer` to `github.com/wlan0/findjson/FindJSON`

```golang
package main

import (
	"strings"
	"fmt"
	"github.com/wlan0/findjson"
	"os"
)

func main() {
	positions, err := findjson.FindJSON(strings.NewReader(`error writing file {"func": "OpenFile", "reason": "file not found", "code": 500}`, os.Stdout))
	if err != nil {
		fmt.Printf("error finding json: %v\n", err)
		return	
	}
	// returns even number of elements
	for i := 0;i<len(positions); i++ {
		fmt.Printf("json document found at [%d, %d]", positions[i].Start, positions[i].End)
	}
}
```
This function call will return a slice of `{start, end}` positions, and write the JSON parts of the text into `io.Writer`

Features
---------

 - [x] Streaming JSON
 - [x] Thread safe
 - [x] GO modules ready
 - [x] Tested

License
---------

Apache V2
