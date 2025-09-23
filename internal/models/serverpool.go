package models

type ServerPool struct {
	Backends     []*Backend
	CurrentIndex uint64
}

func NewServerPool() *ServerPool {
	var backends []*Backend
	return &ServerPool{
		Backends:     backends,
		CurrentIndex: 0,
	}
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.Backends = append(s.Backends, backend)
}

func (sp *ServerPool) GetBackends() []*Backend {
	var backends []*Backend
	for _, b := range sp.Backends {
		if b.IsAlive() {
			backends = append(backends, b)
		}
	}
	return backends
}

func (sp *ServerPool) GetBackendCount() int {
	return len(sp.Backends)
}
