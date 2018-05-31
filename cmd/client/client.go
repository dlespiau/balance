package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

func hostname(url, key string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Affinity", key)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}

func send(url string, key string, n int, delay time.Duration) {
	var wg sync.WaitGroup

	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			name, err := hostname(url, key)
			fmt.Println(name)
			if err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}()
		time.Sleep(delay)
	}

	wg.Wait()
}

func main() {
	url := flag.String("url", "", "proxy url")
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		send(*url, "Marc", 100, 10*time.Millisecond)
		wg.Done()
	}()

	go func() {
		send(*url, "Sophie", 100, 10*time.Millisecond)
		wg.Done()
	}()

	wg.Wait()
}
