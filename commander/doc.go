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

// Package commander converts user input and scripts into commands for gott.
// Many of these commands are undoable. To support undo, the commander creates
// Operation objects that return their inverses when they are performed.
// Commands which are not undoable are implemented directly with calls
// to the editor or other subsystems..
package commander
