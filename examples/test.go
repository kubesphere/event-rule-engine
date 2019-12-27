//for vistor pattern
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wanjunlei/event-rule-engine/visitor"
	"io/ioutil"
	"os"

	"github.com/c-bata/go-prompt"
)

// Flatten takes a map and returns a new one where nested maps are replaced
// by dot-delimited keys.
//func Flatten(key string, v interface{}) map[string]interface{} {
//	o := make(map[string]interface{})
//
//	switch v.(type) {
//	case map[string]interface{}:
//		for mk, mv := range v.(map[string]interface{}) {
//			switch mv.(type) {
//			case map[string]interface{}:
//				nm := Flatten(mk, mv)
//				for nk, nv := range nm {
//					o[mk+"."+nk] = nv
//				}
//			case []interface{}:
//				nm := Flatten(mk, mv)
//				for nk, nv := range nm {
//					o[nk] = nv
//				}
//			default:
//				o[mk] = mv
//			}
//		}
//	case []interface{}:
//		for index,sv :=range v.([]interface{}) {
//			sm := Flatten(key, sv)
//			for nk, nv := range sm {
//				//o[key + "[" + fmt.Sprint(index) + "]"+"."+nk] = nv
//				o[key + "[" + fmt.Sprint(index) + "]"+"."+nk] = nv
//			}
//		}
//	default:
//		o[key] = v
//
//	}
//
//	return o
//}
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
	fmt.Println(visitor.CheckRule(in))
	err, res := visitor.EventRuleEvaluate(fm, in)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Answer: %v\n", res)
}

func completer(_ prompt.Document) []prompt.Suggest {
	var ret []prompt.Suggest
	return ret
}

func readJson() {
	f, err := os.Open("test.json")
	if err != nil {
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
	flag.Parse()
	readJson()
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("EventRuleEngine"),
	)
	p.Run()
}
