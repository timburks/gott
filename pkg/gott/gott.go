
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

package gott

type Editor struct {
	State  int
	Buffer Buffer
}

type Event struct {
	Code     int
	Key      int
	Modifier int
	X        int
	Y        int
}

type Buffer struct {
	Characters string
	Rows       int
	Cols       int
}

func NewEditor() *Editor {
	return &Editor{}
}

func (e *Editor)HandleEvent(ev *Event) *Buffer {
	return &e.Buffer
}
