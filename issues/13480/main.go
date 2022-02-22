package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type data struct {
	Name string `json:"name,omitempty"`
	Age int `json:"age,omitempty"`
	Time time.Duration `json:"time,omitempty"`
}

func main() {
	d1 := data{
		Name: "ahrtr",
		Age: 13,
		Time: time.Second * 40,
	}

	if dataB, err := json.Marshal(d1); err != nil {
		panic (err)
	} else {
		fmt.Println(string(dataB))
	}

	rawStr := `
{
	"name": "ahrtr",
	"age": 13,
	"time": 40000000000
}
`
	var d2 data
	if err := json.Unmarshal([]byte(rawStr), &d2); err != nil {
		panic (err)
	} else {
		fmt.Println(d2)
	}
}

