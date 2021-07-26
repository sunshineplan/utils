package metadata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Server contains metadata server address and verify header.
type Server struct {
	// metadata server address
	Addr string
	// metadata server verify header name
	Header string
	// metadata server verify header value
	Value string
}

// Get queries metadata from the metadata server.
func (s *Server) Get(metadata string, data interface{}) error {
	return s.GetWithClient(
		metadata,
		data,
		&http.Client{Transport: &http.Transport{Proxy: nil}},
	)
}

// GetWithClient queries metadata from the metadata server
// with custom http.Client.
func (s *Server) GetWithClient(metadata string, data interface{}, client *http.Client) error {
	url := s.Addr + "/" + metadata
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to make request to %s: %v", url, err)
	}
	req.Header.Add(s.Header, s.Value)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("no StatusOK response from %s", url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, data)
}
