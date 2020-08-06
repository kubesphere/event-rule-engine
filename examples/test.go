/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/wanjunlei/event-rule-engine/visitor"
	"io/ioutil"
	"os"
)

func Flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := Flatten(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}

var fm map[string]interface{}

func executor(in string) {

	if _, err := visitor.CheckRule(in); err != nil {
		fmt.Printf("rule condition is not correct, %s", err.Error())
		return
	}

	err, res := visitor.EventRuleEvaluate(fm, in)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Answer: %v\n", res)
}

func readJson() {
	f, err := os.Open("examples//test.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	data, _ := ioutil.ReadAll(f)
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		fmt.Println("Unmarshal failed, ", err)
		return
	}

	fm = Flatten(m)
	fmt.Println(fm)
}

func main() {
	readJson()
	executor("ResponseObject.status.images[*].names[*] contains \"kubesphere\"")
}
