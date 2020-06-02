package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	"os"
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

type Entry struct {
	unit     string
	strval   string
	floatval float64
}

func GetData(conf Config, fields []string) map[string]*Entry {
	url := conf.Base + conf.Path + "?apikey=" + conf.ApiKey + "&lat=" + conf.Lat + "&lon=" + conf.Lon + "&unit_system=" + conf.Units
	var n = 0
	for _, name := range fields {
		if n > 0 {
			url = url + "%2C"
		} else {
			url = url + "&fields="
		}

		url = url + name
		n++
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
	result := make(map[string]*Entry)
	for _, val := range fields {
		temp := anyJson[val].(map[string]interface{})
		switch tval := temp["value"].(type) {
		case string:
			if tm, err := time.ParseInLocation(time.RFC3339, tval, loc); err == nil {
				tstr := tm.In(time.Local).Format(time.UnixDate)
				result[val] = &Entry { strval: tstr }
			} else {
				result[val] = &Entry { strval: tval }
			}

		case float64: result[val] = &Entry { strval: "", floatval: tval }
		default: panic("Unknown type in " + val)
		}

		if unit, ok := temp["units"].(string); ok {
			result[val].unit = unit
		} else {
			result[val].unit = ""
		}
	}

	return result
}
