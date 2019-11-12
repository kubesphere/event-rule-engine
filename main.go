//for vistor pattern
package main

import (
	"encoding/json"
	"fmt"
	"event-rule-engine/visitor"
	"io/ioutil"
	"os"

	prompt "github.com/c-bata/go-prompt"
)

// Flatten takes a map and returns a new one where nested maps are replaced
// by dot-delimited keys.
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
	fmt.Printf("Answer: %v\n", visitor.EventRuleEvaluate(fm, in))
}

func completer(in prompt.Document) []prompt.Suggest {
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
	readJson()
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("EventRuleEngine"),
	)
	p.Run()
}
