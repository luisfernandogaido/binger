package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Contador struct {
	sync.RWMutex
	n int
}

func (c *Contador) Incrementa() {
	c.Lock()
	c.n++
	c.Unlock()
}

func (c *Contador) Get() int {
	c.RLock()
	n := c.n
	c.RUnlock()
	return n
}

var (
	replacer  *strings.Replacer
	sucessos  Contador
	fracassos Contador
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
	http.DefaultClient.Timeout = time.Second * 20
	rand.Seed(time.Now().UnixNano())
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
	var (
		q, d, k string
	)
	flag.StringVar(&q, "q", "gatos", "termos de pesquisa")
	flag.StringVar(&d, "d", "./files", "diretório em que as imagens serão armazenadas.")
	flag.StringVar(&k, "k", "080dc190775c4f488a857b2110a4e120", "Bing Search v7 key ")
	flag.Parse()
	if d == "./files" {
		os.Mkdir(d, 0644)
	}
	t0 := time.Now()
	if err := downloadAll(q, d, k); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Executado em %v segundos.\n", time.Since(t0))
}

func downloadAll(q, d, k string) error {
	fmt.Println("Buscando endereços...")
	values, err := search(q, k)
	fmt.Printf("%v endereços retornados. Iniciando download...\n", len(values))
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
			if err == nil {
				sucessos.Incrementa()
				s := sucessos.Get()
				if s%50 == 0 {
					fmt.Println(s)
				}
			} else {
				fracassos.Incrementa()
			}
		}(values[i])
	}
	for i := 0; i < c; i++ {
		sem <- struct{}{}
	}
	return nil
}

func search(q, k string) ([]Value, error) {
	values := make([]Value, 0)
	totalMatches := 10000
	offset := 0
	for offset < totalMatches {
		params := url.Values{
			"offset": []string{strconv.Itoa(offset)},
			"q":      []string{q},
			"count":  []string{"150"},
		}
		req, err := http.NewRequest(
			"GET",
			"https://api.cognitive.microsoft.com/bing/v7.0/images/search?"+params.Encode(),
			nil,
		)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Ocp-Apim-Subscription-Key", k)
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
	b := make([]byte, 2)
	rand.Read(b)
	return ioutil.WriteFile(filepath.Join(dir, replacer.Replace(nam)+fmt.Sprintf(" %x", b)+"."+ext), bytes, 0644)
}
