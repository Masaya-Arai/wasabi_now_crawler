package common

import (
	"sync"
	"net/http"
)

type JobController struct {
	Channel   chan int
	WaitGroup sync.WaitGroup
}

type OptionalString struct {
	String string
	Status bool
}

type HttpChannels struct {
	TransportChannel chan *http.Transport
	RequestChannel   chan *http.Request
}

func (optionalString *OptionalString) Make(str string) *OptionalString {
	optionalString.String = str

	if str == "" {
		optionalString.Status = false
	} else {
		optionalString.Status = true
	}

	return optionalString
}

func (chs *HttpChannels) CloseAll() {
	close(chs.TransportChannel)
	close(chs.RequestChannel)
}