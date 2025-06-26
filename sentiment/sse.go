package sentiment

import (
	"fmt"
	"net/http"
)

type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	return &SSEWriter{w, flusher}, nil
}

func (s *SSEWriter) WriteEvent(data string) error {
	_, err := fmt.Fprintf(s.w, "data: %s\n\n", data)
	s.flusher.Flush()
	return err
}

func (s *SSEWriter) Error(err error) {
	fmt.Fprintf(s.w, "data: [ERROR] %s\n\n", err.Error())
	s.flusher.Flush()
}

func (s *SSEWriter) Done() {
	fmt.Fprint(s.w, "data: [DONE]\n\n")
	s.flusher.Flush()
}
