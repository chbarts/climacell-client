package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"flag"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Base    string   `json:"base"`
	Path    string   `json:"path"`
	Lat     string   `json:"lat"`
	Lon     string   `json:"lon"`
	ApiKey  string   `json:"apikey"`
	Units   string   `json:"unit_system"`
	Options []string `json:"options"`
}

func check(err error) {
        if err != nil {
                panic(err)
        }
}

func ReadConf() Config {
	confdir := os.Getenv("XDG_CONFIG_HOME")
	if len(confdir) == 0 {
		confdir = os.Getenv("HOME")
	}

	path := confdir + "/climacell.json"
	input, err := ioutil.ReadFile(path)
	check(err)

	var res Config
	err = json.Unmarshal([]byte(input), &res)
	check(err)

	return res
}

func main() {
	conf := ReadConf()
	url := conf.Base + conf.Path + "?apikey=" + conf.ApiKey + "&lat=" + conf.Lat + "&lon=" + conf.Lon + "&unit_system=" + conf.Units
	opts := make(map[string]*bool)
	if len(conf.Options) > 0 {
		for _, str := range conf.Options {
			opts[str] = flag.Bool(str, false, "Add this information to result")
		}
	}

	flag.Parse()

	var n = 0
	var clopts []string
	for key, val := range opts {
		if *val {
			if n > 0 {
				url = url + "%2C"
			} else {
				url = url + "&fields="
			}

			url = url + key
			clopts = append(clopts, key)
			n++
		}
	}

	res, err := http.Get(url)
	check(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	check(err)

	// https://stackoverflow.com/questions/42152750/golang-is-there-an-easy-way-to-unmarshal-arbitrary-complex-json
	var anyJson map[string]interface{}
	json.Unmarshal(body, &anyJson)

	loc := time.Now().Location()
	for _, val := range clopts {
		temp := anyJson[val].(map[string]interface{})
		switch tval := temp["value"].(type) {
		case string:
			if tm, err := time.ParseInLocation(time.RFC3339, tval, loc); err == nil {
				fmt.Printf("%s = %s ", val, tm.In(time.Local).Format(time.UnixDate))
			} else {
				fmt.Printf("%s = %s ", val, tval)
			}

		case float64: fmt.Printf("%s = %f ", val, tval)
		default: panic("Unknown type in " + val)
		}

		if unit, ok := temp["units"].(string); ok {
			fmt.Printf("%s", unit)
		}

		fmt.Printf("\n")
	}
}
