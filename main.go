package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type waiter struct {
	ch chan string
}

type queue struct {
	mu       sync.Mutex
	messages []string
	waiters  []*waiter
}

var (
	queuesMu sync.Mutex
	queues   = make(map[string]*queue)
)

func getQueue(name string) *queue {
	queuesMu.Lock()
	defer queuesMu.Unlock()
	if q, ok := queues[name]; ok {
		return q
	}
	q := &queue{}
	queues[name] = q
	return q
}

func (q *queue) put(msg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.waiters) > 0 {
		w := q.waiters[0]
		q.waiters = q.waiters[1:]
		w.ch <- msg
		return
	}
	q.messages = append(q.messages, msg)
}

func (q *queue) get(timeout time.Duration) (string, bool) {
	q.mu.Lock()
	if len(q.messages) > 0 {
		msg := q.messages[0]
		q.messages = q.messages[1:]
		q.mu.Unlock()
		return msg, true
	}
	if timeout == 0 {
		q.mu.Unlock()
		return "", false
	}
	w := &waiter{ch: make(chan string, 1)}
	q.waiters = append(q.waiters, w)
	q.mu.Unlock()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case msg := <-w.ch:
		return msg, true
	case <-timer.C:
		q.mu.Lock()
		for i, x := range q.waiters {
			if x == w {
				q.waiters = append(q.waiters[:i], q.waiters[i+1:]...)
				break
			}
		}
		q.mu.Unlock()
		select {
		case msg := <-w.ch:
			return msg, true
		default:
			return "", false
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	q := getQueue(name)

	switch r.Method {
	case http.MethodPut:
		msg := r.URL.Query().Get("v")
		if msg == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		q.put(msg)
		w.WriteHeader(http.StatusOK)

	case http.MethodGet:
		timeout := time.Duration(0)
		if s := r.URL.Query().Get("timeout"); s != "" {
			sec, err := strconv.Atoi(s)
			if err != nil || sec < 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			timeout = time.Duration(sec) * time.Second
		}
		if msg, ok := q.get(timeout); ok {
			fmt.Fprint(w, msg)
			return
		}
		w.WriteHeader(http.StatusNotFound)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <port>\n", os.Args[0])
		os.Exit(1)
	}
	addr := ":" + os.Args[1]
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
