package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sjnam/oproc"
)

type item struct {
	ID      string `json:"id,omitempty"`
	Key     string `json:"key,omitempty"`
	Name    string `json:"name,omitempty"`
	Caption string `json:"caption,omitempty"`
	Authors []struct {
		ID     string `json:"id"`
		Active bool   `json:"active"`
	} `json:"authors,omitempty"`
	Topics []string `json:"topics,omitempty"`
}

func main() {
	inputStream := func() <-chan item {
		valStream := make(chan item)
		go func() {
			defer close(valStream)
			resp, err := http.Get("https://ctan.org/json/2.0/packages")
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			var list []item
			if err = json.NewDecoder(resp.Body).Decode(&list); err != nil {
				log.Fatal(err)
			}

			for _, o := range list {
				valStream <- o
			}
		}()
		return valStream
	}

	doWork := func(o item) item {
		resp, err := http.Get("https://ctan.org/json/2.0/pkg/" + o.Key)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&o); err != nil {
			log.Fatal(o.ID, err)
		}
		return o
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for s := range oproc.OrderedProc(ctx, inputStream(), doWork, 20) {
		b, err := json.Marshal(s)
		if err == nil {
			fmt.Println(string(b))
		}
	}
}
