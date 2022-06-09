package main

import (
	"sort"
	"sync"
)

type RpcEndpointPool []RpcEndpoint

var RetryMutex sync.Mutex

type RpcEndpoint struct {
	Piority int
	Host    string
	Retry   int
}

func (r RpcEndpointPool) Len() int           { return len(r) }
func (r RpcEndpointPool) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r RpcEndpointPool) Less(i, j int) bool { return r[i].Piority < r[j].Piority }

func sortEndpoint(endpoints []RpcEndpoint) {
	sort.Sort(RpcEndpointPool(endpoints))
}

func (r *RpcEndpoint) GoNext(max_retry int) bool {
	if r.Retry < max_retry {
		return false
	}
	return true
}

func (r *RpcEndpoint) AddRetry() {
	RetryMutex.Lock()
	defer RetryMutex.Unlock()
	r.Retry += 1

}

func (r *RpcEndpoint) ResetRetry() {
	RetryMutex.Lock()
	defer RetryMutex.Unlock()
	r.Retry = 0
}
