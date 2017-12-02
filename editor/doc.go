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

// Package editor implements the core text editing functions of gott.
// Many of these functions are accessed only through operations; this
// makes it possible to easily repeat and undo them.
// An editor manages multiple windows; windows are rectangular 
// subdivisions of a screen and each window edits an associated
// buffer. It is possible for multiple windows to concurrently edit
// a single buffer.
package editor
