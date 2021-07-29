package agamemnon

import (
	"sync"
)

type RequestQueue struct {
	sync.RWMutex
	// Design Idea - key: client IP-Port hash, value: RequestCache key
	queue []string
}

var requestQueue RequestQueue

func (r *RequestQueue) Init()  {
	r.Lock()
	defer r.Unlock()
	r.queue = []string{}
}

func (r *RequestQueue) Enqueue(val string)  {
	r.Lock()
	defer r.Unlock()
	r.queue = append(r.queue, val)
}

func (r *RequestQueue) Deque() (string, bool) {
	r.Lock()
	defer r.Unlock()

	if r.size() == 0 {
		return "", false
	}

	res := r.queue[0]
	r.queue = r.queue[1:]
	return res, true
}

func (r *RequestQueue) Peek() (string, bool) {
	r.RLock()
	defer r.RUnlock()

	if r.size() == 0 {
		return "", false
	}
	return r.queue[0], true
}

func (r *RequestQueue) size() int {
	return len(r.queue)
}
