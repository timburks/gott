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
package commander

import (
	"errors"
	"log"

	"github.com/steelseries/golisp"
)

func init() {
	golisp.Global.BindTo(golisp.SymbolWithName("CONSTANT"), golisp.FloatWithValue(float32(2.0)))
	golisp.MakePrimitiveFunction("go-fact", "1", GoFactImpl)
}

func GoFactImpl(args *golisp.Data, env *golisp.SymbolTableFrame) (result *golisp.Data, err error) {
	val := golisp.Car(args)
	if !golisp.FloatP(val) {
		return nil, errors.New("go-fact requires a float argument")
	}
	n := int(golisp.FloatValue(val))
	f := 1
	for i := 1; i <= n; i++ {
		f *= i
	}
	return golisp.FloatWithValue(float32(f)), nil
}

func ParseEval(command string) {
	value, err := golisp.ParseAndEval(command)
	if err != nil {
		log.Printf("ERR %+v", err)
	} else {
		log.Printf("SEXPR %+v", value)
		if golisp.FloatP(value) {
			log.Printf("VALUE %+v", golisp.FloatValue(value))
		}
	}
}
