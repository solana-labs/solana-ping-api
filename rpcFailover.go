package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/portto/solana-go-sdk/client"
)

var failoverMutex sync.Mutex

type FailoverEndpointList []FailoverEndpoint
type FailoverEndpoint struct {
	Endpoint    string
	AccessToken string
	Piority     int
	MaxRetry    int
	Retry       int
}

type RPCFailover struct {
	curIndex  int
	Endpoints []FailoverEndpoint
}

func (f FailoverEndpointList) Len() int           { return len(f) }
func (f FailoverEndpointList) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f FailoverEndpointList) Less(i, j int) bool { return f[i].Piority < f[j].Piority }

func NewRPCFailover(endpoints []RPCEndpoint) RPCFailover {
	fList := []FailoverEndpoint{}
	for _, e := range endpoints {
		endpointExist := strings.TrimRight(e.Endpoint, " /")
		if len(endpointExist) == 0 {
			continue
		}
		tokenExist := strings.Trim(e.AccessToken, " ")
		fList = append(fList,
			FailoverEndpoint{
				Endpoint:    endpointExist,
				AccessToken: tokenExist,
				Piority:     e.Piority,
				MaxRetry:    e.MaxRetry})
	}
	sort.Sort(FailoverEndpointList(FailoverEndpointList(fList)))
	return RPCFailover{
		Endpoints: fList,
	}
}

func (f *RPCFailover) IsFail() bool {
	if f.Endpoints[f.curIndex].Retry >= f.Endpoints[f.curIndex].MaxRetry {
		return true
	}
	return false
}

func (f *RPCFailover) GetNext() string {
	failoverMutex.Lock()
	defer failoverMutex.Unlock()
	f.curIndex = f.curIndex + 1
	return f.Endpoints[f.curIndex].Endpoint
}

func (f *RPCFailover) GoNext(cur *client.Client, config ClusterConfig, workerNum int) *client.Client {
	failoverMutex.Lock()
	defer failoverMutex.Unlock()
	var next *client.Client
	retries := f.GetEndpoint().Retry
	if retries < f.GetEndpoint().MaxRetry { // Go Next
		if cur != nil {
			return cur
		} else {
			connectionEndpoint := f.GetEndpoint().Endpoint
			if len(f.GetEndpoint().AccessToken) != 0 {
				connectionEndpoint = fmt.Sprintf("%s/%s", f.GetEndpoint().Endpoint, f.GetEndpoint().AccessToken)
			}
			return client.NewClient(connectionEndpoint)
		}
	}
	idx := f.GetNextIndex()
	if len(f.Endpoints[idx].AccessToken) != 0 {
		next = client.NewClient(fmt.Sprintf("%s/%s", f.Endpoints[idx].Endpoint, f.Endpoints[idx].AccessToken))
	} else {
		next = client.NewClient(f.Endpoints[idx].Endpoint)
	}
	log.Println("GoNext!!! New Endpoint:", f.GetEndpoint())
	if config.AlternativeEnpoint.SlackAlert.Enabled {
		var slack SlackPayload
		slack.FailoverAlertPayload(config, *f.GetEndpoint(), workerNum)
		SlackSend(config.AlternativeEnpoint.SlackAlert.Webhook, &slack)
	}
	return next
}

func (f *RPCFailover) GetEndpoint() *FailoverEndpoint {
	return &f.Endpoints[f.curIndex]
}

func (f *FailoverEndpoint) RetryResult(err PingResultError) {
	failoverMutex.Lock()
	defer failoverMutex.Unlock()
	if !err.NoError() {
		if err.IsTooManyRequest429() ||
			err.IsServiceUnavilable() ||
			err.IsErrGatewayTimeout504() ||
			err.IsNoSuchHost() {
			f.Retry += 1
		}
	} else {
		f.Retry = 0
	}
}

func (f *RPCFailover) GetNextIndex() int {
	if f.curIndex < 0 {
		log.Panic("current Index of FailoverEndpoint < 0")
	}
	if f.curIndex+1 > len(f.Endpoints)-1 {
		f.curIndex = 0
	} else {
		f.curIndex = f.curIndex + 1
	}
	return f.curIndex
}
