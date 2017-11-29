//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package editor

import (
	"encoding/hex"
	"regexp"

	gott "github.com/timburks/gott/types"
)

// The GoHighlighter highlights Go code.
type GoHighlighter struct {
	hexPattern          *regexp.Regexp
	punctuationPattern  *regexp.Regexp
	commentPattern      *regexp.Regexp
	quotedStringPattern *regexp.Regexp
	keywordPattern      *regexp.Regexp
	numberPattern       *regexp.Regexp
}

func NewGoHighlighter() *GoHighlighter {
	h := &GoHighlighter{}

	h.hexPattern, _ = regexp.Compile("0x[0-9|a-f][0-9|a-f]")
	h.punctuationPattern, _ = regexp.Compile("\\(|\\)|,|:|=|\\[|\\]|\\{|\\}|\\+|-|\\*|<|>|;")
	h.commentPattern, _ = regexp.Compile("\\/\\/.*$")
	h.quotedStringPattern, _ = regexp.Compile("\"[^\"]*\"")
	h.keywordPattern, _ = regexp.Compile("break|default|func|interface|select|case|defer|go|map|struct|chan|else|goto|package|switch|const|fallthrough|if|range|type|continue|for|import|return|var")
	h.keywordPattern.Longest()
	h.numberPattern, _ = regexp.Compile("([0-9]+(\\.[0-9]*)?)|(([0-9]*\\.)?[0-9]+)")

	return h
}

func (h *GoHighlighter) Highlight(b *Buffer) {

	for _, r := range b.rows {

		colors := r.GetColors()

		for j, _ := range colors {
			colors[j] = 0xff
		}

		line := string(r.GetText())
		matches := h.keywordPattern.FindAllSubmatchIndex([]byte(line), -1)
		if matches != nil {
			for _, match := range matches {
				// if there's an alphanumeric character on either side, skip this
				if !checkalphanum(line, match[0], match[1]) {
					for k := match[0]; k < match[1]; k++ {
						colors[k] = 0x70
					}
				}
			}
		}

		matches = h.numberPattern.FindAllSubmatchIndex([]byte(line), -1)
		if matches != nil {
			for _, match := range matches {
				// if there's an alphanumeric character on either side, skip this
				if !checkalphanum(line, match[0], match[1]) {
					for k := match[0]; k < match[1]; k++ {
						colors[k] = 0x83
					}
				}
			}
		}

		matches = h.punctuationPattern.FindAllSubmatchIndex([]byte(line), -1)
		if matches != nil {
			for _, match := range matches {
				for k := match[0]; k < match[1]; k++ {
					colors[k] = 0x71
				}
			}
		}

		matches = h.hexPattern.FindAllSubmatchIndex([]byte(line), -1)
		if matches != nil {
			for _, match := range matches {
				for k := match[0]; k < match[1]; k++ {
					x, _ := hex.DecodeString(line[match[0]+2 : match[1]])
					colors[k] = gott.Color(x[0])
				}
			}
		}
		matches = h.quotedStringPattern.FindAllSubmatchIndex([]byte(line), -1)
		if matches != nil {
			for _, match := range matches {
				for k := match[0]; k < match[1]; k++ {
					colors[k] = 0xe0
				}
			}
		}
		matches = h.commentPattern.FindAllSubmatchIndex([]byte(line), -1)
		if matches != nil {
			for _, match := range matches {
				for k := match[0]; k < match[1]; k++ {
					colors[k] = 0xf8
				}
			}
		}
	}

}
