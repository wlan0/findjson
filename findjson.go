/*
 *  Copyright 2019 Sidhartha Mani
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
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

package findjson

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const (
	parseSuccess = iota
	parseNext
	parseError
)

type Position struct {
	Start int
	End   int
}

func FindJSON(dataStream io.Reader, outStream io.Writer) ([]Position, error) {
	g := []rune{}
	pos := []int{}
	rewind := false
	b := bufio.NewReader(dataStream)

	var err error
	var r rune
	var index int
	var state int
	var parseFnInt interface{}
	parseFn := parseJSON
	positions := []Position{}

	rw := &bytes.Buffer{}
	inBlock := false
	for {
		if !rewind {
			r, _, err = b.ReadRune()
			if err != nil {
				if err == io.EOF {
					if len(pos)%2 != 0 && len(g) == 0 && len(pos) > 1 {
						pos = append(pos, index-1)
						positions = append(positions, Position{
							Start: pos[len(pos)-2],
							End:   pos[len(pos)-1],
						})
						rw.WriteRune('\n')
						_, err = rw.WriteTo(outStream)
						if err != nil {
							return positions, fmt.Errorf("Error writing to output stream: %v", err)
						}
					}
					return positions, nil
				}
				return nil, fmt.Errorf("error reading stream: %v\n", err)
			}
			if inBlock {
				rw.WriteRune(r)
			}
		}
		rewind = false
		g, parseFnInt, rewind, state = parseFn(g, r)

		parseFn = parseFnInt.(func([]rune, rune) ([]rune, interface{}, bool, int))

		if len(pos)%2 == 0 && (r == '{' || r == '[') {
			inBlock = true
			rw.WriteRune(r)
			pos = append(pos, index)
		}
		switch state {
		case parseSuccess:
			pos = append(pos, index)
			parseFn = parseJSON
			if len(pos) > 1 {
				positions = append(positions, Position{
					Start: pos[len(pos)-2],
					End:   pos[len(pos)-1],
				})
			}
			rw.WriteRune('\n')
			_, err = rw.WriteTo(outStream)
			if err != nil {
				fmt.Printf("Error writing to output stream: %v\n", err)
			}
			rw.Reset()
			inBlock = false
		case parseError:
			if len(pos)%2 != 0 {
				pos = pos[:len(pos)-1]
			}
			rw.Reset()
			inBlock = false
			parseFn = parseJSON
			g = []rune{}
		case parseNext:
		}
		index++
	}
	return positions, nil
}

func parseJSON(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
		return state, parseJSON, false, parseNext
	}
	if r == '{' {
		state = append(state, '{')
		parseFn := parseJSONCurlyStart
		return state, parseFn, false, parseNext
	}
	if r == '[' {
		state = append(state, '[')
		return state, parseJSON, false, parseNext
	}
	if len(state) == 0 {
		return state, parseJSON, false, parseError
	}
	if r == ']' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, false, parseNext
	}
	if r == '}' {
		if state[len(state)-1] != '{' {
			return state, parseJSON, false, parseError
		}
		state = state[:len(state)-1]
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, false, parseNext
	}
	if r == '"' {
		parseFn := parseJSONString
		return state, parseFn, false, parseNext
	}
	if r > '0' && r <= '9' {
		parseFn := parseJSONIntOrFloat
		return state, parseFn, false, parseNext
	}
	if r == '0' {
		parseFn := parseJSONFloatOrZero
		return state, parseFn, false, parseNext
	}
	if r == 't' {
		parseFn := parseJSONBoolT
		return state, parseFn, false, parseNext
	}
	if r == 'f' {
		parseFn := parseJSONBoolF
		return state, parseFn, false, parseNext
	}
	if r == 'n' {
		parseFn := parseJSONN
		return state, parseFn, false, parseNext
	}
	if r == '-' {
		parseFn := parseJSONNegativeNum
		return state, parseFn, false, parseNext
	}
	return state, parseJSON, false, parseError
}

func parseJSONN(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'u' {
		parseFn := parseJSONNu
		return state, parseFn, false, parseNext
	}
	return state, parseJSONN, false, parseError
}

func parseJSONNu(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'l' {
		parseFn := parseJSONNul
		return state, parseFn, false, parseNext
	}
	return state, parseJSONNu, false, parseError
}

func parseJSONNul(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'l' {
		parseFn := parseJSONNull
		return state, parseFn, false, parseNext
	}
	return state, parseJSONNul, false, parseError
}

func parseJSONNull(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' || r == '}' || r == ']' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r == ',' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	return state, parseJSONNull, false, parseError
}

func parseJSONBoolT(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'r' {
		parseFn := parseJSONBoolTr
		return state, parseFn, false, parseNext
	}
	return state, parseJSONBoolT, false, parseError
}

func parseJSONBoolTr(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'u' {
		parseFn := parseJSONBoolTru
		return state, parseFn, false, parseNext
	}
	return state, parseJSONBoolTr, false, parseError
}

func parseJSONBoolTru(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'e' {
		parseFn := parseJSONBoolTrue
		return state, parseFn, false, parseNext
	}
	return state, parseJSONBoolTru, false, parseError
}

func parseJSONBoolTrue(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' || r == '}' || r == ']' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r == ',' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	return state, parseJSONBoolTrue, false, parseError
}

func parseJSONBoolF(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'a' {
		parseFn := parseJSONBoolFa
		return state, parseFn, false, parseNext
	}
	return state, parseJSONBoolF, false, parseError
}

func parseJSONBoolFa(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'l' {
		parseFn := parseJSONBoolFal
		return state, parseFn, false, parseNext
	}
	return state, parseJSONBoolFa, false, parseError
}

func parseJSONBoolFal(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 's' {
		parseFn := parseJSONBoolFals
		return state, parseFn, false, parseNext
	}
	return state, parseJSONBoolFal, false, parseError
}

func parseJSONBoolFals(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == 'e' {
		parseFn := parseJSONBoolFalse
		return state, parseFn, false, parseNext
	}
	return state, parseJSONBoolFals, false, parseError
}

func parseJSONBoolFalse(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' || r == '}' || r == ']' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r == ',' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	return state, parseJSONBoolFalse, false, parseError
}

func parseJSONNegativeNum(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r <= '9' && r > '0' {
		return state, parseJSONIntOrFloat, false, parseNext
	}
	return state, parseJSONNegativeNum, false, parseError
}

func parseJSONIntOrFloat(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' || r == '}' || r == ']' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r == ',' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}

	if r <= '9' && r >= '0' {
		return state, parseJSONIntOrFloat, false, parseNext
	}
	if r == '.' {
		parseFn := parseJSONFloatRestFirst
		return state, parseFn, false, parseNext
	}
	if r == 'E' || r == 'e' {
		return state, parseJSONExpRestNumOrSign, false, parseNext
	}
	return state, parseJSONIntOrFloat, false, parseError
}

func parseJSONFloatRestFirst(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r <= '9' && r >= '0' {
		parseFn := parseJSONFloatRest
		return state, parseFn, false, parseNext
	}
	return state, parseJSONFloatRestFirst, false, parseError
}

func parseJSONFloatRest(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' || r == '}' || r == ']' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r == ',' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r <= '9' && r >= '0' {
		return state, parseJSONFloatRest, false, parseNext
	}
	if r == 'E' || r == 'e' {
		return state, parseJSONExpRestNumOrSign, false, parseNext
	}
	return state, parseJSONFloatRest, false, parseError
}

func parseJSONExpRestNumOrSign(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r >= '1' && r <= '9' {
		return state, parseJSONExpRestRest, false, parseNext
	}
	if r == '+' || r == '-' {
		return state, parseJSONExpRestSign, false, parseNext
	}
	return state, parseJSONExpRestNumOrSign, false, parseError
}

func parseJSONExpRestSign(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r >= '1' && r <= '9' {
		return state, parseJSONExpRestRest, false, parseNext
	}
	return state, parseJSONExpRestSign, false, parseError
}

func parseJSONExpRestRest(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' || r == '}' || r == ']' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r <= '9' && r >= '0' {
		return state, parseJSONExpRestRest, false, parseNext
	}
	if r == ',' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	return state, parseJSONExpRestRest, false, parseError
}

func parseJSONFloatOrZero(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' || r == '}' || r == ']' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r == ',' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r == '.' {
		parseFn := parseJSONFloatRestFirst
		return state, parseFn, false, parseNext
	}
	return state, parseJSONFloatOrZero, false, parseError
}

func parseJSONCurlyStart(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' {
		return state, parseJSONCurlyStart, false, parseNext
	}
	if r == '}' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, true, parseNext
	}
	if r == '"' {
		parseFn := parseJSONStringKey
		return state, parseFn, false, parseNext
	}
	return state, parseJSONCurlyStart, false, parseError
}

func parseJSONCurlyNext(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == '\n' || r == ' ' || r == '\t' || r == '\r' {
		return state, parseJSONCurlyNext, false, parseNext
	}
	if r == '"' {
		parseFn := parseJSONStringKey
		return state, parseFn, false, parseNext
	}
	return state, parseJSONCurlyNext, false, parseError

}

func parseJSONStringKey(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == ' ' || r == '\t' || r == '\r' {
		return state, parseJSONStringKey, false, parseNext
	}
	if r == '\n' {
		return state, parseJSONStringKey, false, parseError
	}
	if r == '\\' {
		parseFn := parseJSONStringKeyEscapeSeq
		return state, parseFn, false, parseNext
	}
	if r == '"' {
		parseFn := parseJSONKeyEnd
		return state, parseFn, false, parseNext
	}
	return state, parseJSONStringKey, false, parseNext
}

func parseJSONKeyEnd(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == ' ' || r == '\t' || r == '\r' {
		return state, parseJSONKeyEnd, false, parseNext
	}
	if r == ':' {
		parseFn := parseJSON
		return state, parseFn, false, parseNext
	}
	return state, parseJSONKeyEnd, false, parseError
}

func parseJSONStringKeyEscapeSeq(state []rune, r rune) ([]rune, interface{}, bool, int) {
	parseFn := parseJSONStringKey
	return state, parseFn, false, parseNext
}

func parseJSONEndOrNextElement(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
		return state, parseJSONEndOrNextElement, false, parseNext
	}
	if r == ',' {
		var parseFn interface{}
		if state[len(state)-1] == '{' {
			parseFn = parseJSONCurlyNext
		}
		if state[len(state)-1] == '[' {
			parseFn = parseJSON
		}
		return state, parseFn, false, parseNext
	}
	validEnd := false
	if r == ']' {
		if state[len(state)-1] != '[' {
			return state, parseJSONEndOrNextElement, false, parseError
		}
		state = state[:len(state)-1]
		validEnd = true
	}
	if r == '}' {
		if state[len(state)-1] != '{' {
			return state, parseJSONEndOrNextElement, false, parseError
		}
		state = state[:len(state)-1]
		validEnd = true
	}
	if len(state) == 0 {
		return state, parseJSONEndOrNextElement, false, parseSuccess
	}
	if validEnd {
		return state, parseJSONEndOrNextElement, false, parseNext
	}
	return state, parseJSONEndOrNextElement, false, parseError
}

func parseJSONString(state []rune, r rune) ([]rune, interface{}, bool, int) {
	if r == ' ' || r == '\t' || r == '\r' {
		return state, parseJSONString, false, parseNext
	}
	if r == '\n' {
		return state, parseJSONString, false, parseError
	}
	if r == '\\' {
		parseFn := parseJSONStringEscapeSeq
		return state, parseFn, false, parseNext
	}
	if r == '"' {
		parseFn := parseJSONEndOrNextElement
		return state, parseFn, false, parseNext
	}
	return state, parseJSONString, false, parseNext
}

func parseJSONStringEscapeSeq(state []rune, r rune) ([]rune, interface{}, bool, int) {
	parseFn := parseJSONString
	return state, parseFn, false, parseNext
}
