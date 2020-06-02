package main

import (
	"flag"
	"fmt"
)

func main() {
	conf := ReadConf()
	opts := make(map[string]*bool)
	if len(conf.Options) > 0 {
		for _, str := range conf.Options {
			opts[str] = flag.Bool(str, false, "Add this information to result")
		}
	}

	flag.Parse()

	var clopts []string
	for key, val := range opts {
		if *val {
			clopts = append(clopts, key)
		}
	}

	data := GetData(conf, clopts)

	for opt, value := range data {
		fmt.Println(opt + " = " + value)
	}
}
