package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func putURL(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func TestSpecWalkthrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	put := func(path string) int {
		resp, err := putURL(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if len(body) != 0 {
			t.Fatalf("PUT %s: expected empty body, got %q", path, body)
		}
		return resp.StatusCode
	}
	get := func(path string) (int, string) {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}

	if code := put("/pet?v=cat"); code != 200 {
		t.Fatalf("PUT cat: want 200, got %d", code)
	}
	if code := put("/pet?v=dog"); code != 200 {
		t.Fatalf("PUT dog: want 200, got %d", code)
	}
	put("/role?v=manager")
	put("/role?v=executive")

	if code, body := get("/pet"); code != 200 || body != "cat" {
		t.Fatalf("GET pet #1: want 200 cat, got %d %q", code, body)
	}
	if code, body := get("/pet"); code != 200 || body != "dog" {
		t.Fatalf("GET pet #2: want 200 dog, got %d %q", code, body)
	}
	if code, body := get("/pet"); code != 404 || body != "" {
		t.Fatalf("GET pet #3: want 404 empty, got %d %q", code, body)
	}
	if code, body := get("/role"); code != 200 || body != "manager" {
		t.Fatalf("GET role #1: want 200 manager, got %d %q", code, body)
	}
	if code, body := get("/role"); code != 200 || body != "executive" {
		t.Fatalf("GET role #2: want 200 executive, got %d %q", code, body)
	}
	if code, body := get("/role"); code != 404 || body != "" {
		t.Fatalf("GET role #3: want 404 empty, got %d %q", code, body)
	}
}

func TestPutMissingV(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	resp, err := putURL(srv.URL + "/pet")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 400 || len(body) != 0 {
		t.Fatalf("want 400 empty, got %d body=%q", resp.StatusCode, body)
	}
}

func TestGetTimeoutReceivesPut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	done := make(chan string, 1)
	go func() {
		resp, err := http.Get(srv.URL + "/wait?timeout=3")
		if err != nil {
			t.Error(err)
			done <- ""
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		done <- string(body)
	}()

	time.Sleep(200 * time.Millisecond)
	putURL(srv.URL + "/wait?v=hello")

	select {
	case body := <-done:
		if body != "hello" {
			t.Fatalf("want hello, got %q", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GET did not return in time")
	}
}

func TestGetTimeoutEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/empty?timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 404 || len(body) != 0 {
		t.Fatalf("want 404 empty, got %d body=%q", resp.StatusCode, body)
	}
}

func TestWaiterFIFO(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	var wg sync.WaitGroup
	results := make([]string, 2)

	fetch := func(i int) {
		defer wg.Done()
		resp, err := http.Get(srv.URL + "/race?timeout=5")
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		results[i] = string(body)
	}

	wg.Add(2)
	go fetch(0)
	time.Sleep(50 * time.Millisecond)
	go fetch(1)
	time.Sleep(100 * time.Millisecond)

	putURL(srv.URL + "/race?v=first")
	time.Sleep(50 * time.Millisecond)
	putURL(srv.URL + "/race?v=second")

	wg.Wait()

	if results[0] != "first" || results[1] != "second" {
		t.Fatalf("want first/second, got %q / %q", results[0], results[1])
	}
}
