package models

import (
	"log/slog"
	"net/url"
	"sync"
)

type Backend struct {
	URL               *url.URL
	Alive             bool
	ActiveConnections int
	mux               sync.RWMutex
}

func NewBackend(rawURL string) (*Backend, error) {
	parsed_url, err := url.Parse(rawURL)
	if err != nil {
		slog.Error("Error while parsing url", "err", err)
	}
	return &Backend{
		URL:               parsed_url,
		Alive:             true,
		ActiveConnections: 0,
	}, nil
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.Alive
	b.mux.RUnlock()
	return alive
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IncrementConnections() {
	b.mux.Lock()
	b.ActiveConnections += 1
	b.mux.Unlock()
}

func (b *Backend) DecrementConnections() {
	b.mux.Lock()
	b.ActiveConnections -= 1
	b.mux.Unlock()
}

func (b *Backend) GetActiveConnections() int {
	b.mux.Lock()
	c := b.ActiveConnections
	b.mux.Unlock()
	return c
}

//TODO:
/* (b *Backend) UpdateProxy(rawURL string, transport *http.Transport) error */
