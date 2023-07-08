package esbuildfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var ErrFlasherUnsupported = errors.New("the flusher is not supported")

type serverSentEvent struct {
	event string
	data  string
}

type changeEvent struct {
	Update []string `json:"updated"`
}

type ServerSentEventHandler struct {
	mux     sync.RWMutex
	streams []chan serverSentEvent
}

func NewSSE() *ServerSentEventHandler {
	return &ServerSentEventHandler{}
}

func (s *ServerSentEventHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := convertToSSE(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 - event stream error"))
		return
	}

	closed := make(chan bool, 1)

	stream := s.NewStream()
	go serveStream(stream, closed, w)

	select {
	case <-req.Context().Done():
	case <-closed:
	}

	s.RemoveStream(stream)
}

func (s *ServerSentEventHandler) Broadcast(event, data string) {
	for idx := range s.streams {
		s.streams[idx] <- serverSentEvent{event, data}
	}
}

func (s *ServerSentEventHandler) NotifyChanged(updated []string) error {
	data, err := json.Marshal(changeEvent{updated})
	if err != nil {
		return err
	}

	s.Broadcast("change", string(data))
	return nil
}

func (s *ServerSentEventHandler) NewStream() chan serverSentEvent {
	s.mux.Lock()
	defer s.mux.Unlock()

	stream := make(chan serverSentEvent)
	s.streams = append(s.streams, stream)

	return stream
}

func (s *ServerSentEventHandler) RemoveStream(stream chan serverSentEvent) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for idx := range s.streams {
		if s.streams[idx] != stream {
			continue
		}

		end := len(s.streams) - 1
		s.streams[idx] = s.streams[end]
		s.streams = s.streams[:end]

		close(stream)
		break
	}
}

func serveStream(stream chan serverSentEvent, closed chan bool, w http.ResponseWriter) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	for {
		var msg []byte
		select {
		case next, ok := <-stream:
			if !ok {
				closed <- true
				return
			}
			msg = []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", next.event, next.data))
		case <-time.After(30 * time.Second):
			msg = []byte(":\n\n")
		}

		if _, err := w.Write(msg); err != nil {
			return
		}

		flusher.Flush()
	}
}

func convertToSSE(w http.ResponseWriter) (err error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return ErrFlasherUnsupported
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("retry: 500\n"))
	if err != nil {
		return err
	}

	flusher.Flush()
	return nil
}
