//https://dev.cognitive.microsoft.com/docs/services/56b43f0ccf5ff8098cef3808/operations/571fab09dbe2d933e891028f
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	replacer *strings.Replacer
)

func init() {
	replacer = strings.NewReplacer(
		"\\", "",
		"/", "",
		":", "",
		"*", "",
		"?", "",
		"<", "",
		">", "",
		"|", "",
	)
}

type Result struct {
	TotalEstimatedMatches int     `json:"totalEstimatedMatches"`
	NextOffset            int     `json:"nextOffset"`
	Value                 []Value `json:"value"`
}

type Value struct {
	Name           string `json:"name"`
	ContentUrl     string `json:"contentUrl"`
	EncodingFormat string `json:"encodingFormat"`
}

func main() {
	if err := downloadAll("chess", "./files"); err != nil {
		log.Fatal(err)
	}
}

func downloadAll(q, d string) error {
	values, err := search(q)
	if err != nil {
		return err
	}
	c := 128
	sem := make(chan struct{}, c)
	for i := 0; i < len(values); i++ {
		sem <- struct{}{}
		go func(v Value) {
			defer func() { <-sem }()
			err := download(v.ContentUrl, v.Name, v.EncodingFormat, d)
			fmt.Println(v.Name, err)
		}(values[i])
	}
	for i := 0; i < c; i++ {
		sem <- struct{}{}
	}
	return nil
}

func search(q string) ([]Value, error) {
	values := make([]Value, 0)
	totalMatches := 10000
	offset := 0
	for offset < totalMatches {
		params := url.Values{
			"offset": []string{strconv.Itoa(offset)},
			"q":      []string{q},
		}
		req, err := http.NewRequest(
			"GET",
			"https://api.cognitive.microsoft.com/bing/v7.0/images/search?"+params.Encode(),
			nil,
		)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Ocp-Apim-Subscription-Key", "080dc190775c4f488a857b2110a4e120")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		res.Body.Close()
		result := Result{}
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return nil, err
		}
		totalMatches = result.TotalEstimatedMatches
		offset = result.NextOffset
		values = append(values, result.Value...)
		fmt.Println(offset, totalMatches)
	}
	return values, nil
}

func download(end, nam, ext, dir string) error {
	res, err := http.Get(end)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return ioutil.WriteFile(filepath.Join(dir, replacer.Replace(nam)+"."+ext), bytes, 0644)
}
