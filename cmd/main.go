/*
 *   Copyright 2019 Sidhartha Mani
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/wlan0/findjson"
)

func main() {
	files := []io.Reader{}
	if len(os.Args) > 1 {
		for i := range os.Args {
			if i == 0 {
				continue
			}
			f, err := os.Open(os.Args[i])
			if err != nil {
				fmt.Printf("could not open %s: %v", os.Args[i], err)
				os.Exit(1)
			}
			files = append(files, f)
		}
	} else {
		files = append(files, os.Stdin)
	}
	for i, f := range files {
		pos, err := findjson.FindJSON(f, os.Stdout)
		if err != nil {
			fmt.Printf("error finding json in file %d\n", i)
			continue
		}
		if len(pos) == 0 {
			//fmt.Println("no json documents found")
			continue
		}
		//fmt.Printf("json found at %+v\n", pos)
	}
}
