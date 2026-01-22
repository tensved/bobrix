package ctx

import "sync"

type handlesStatus struct {
	isHandled bool
	mx        sync.Mutex
}

func (s *handlesStatus) IsHandledWithUnlocker() (bool, func()) {
	s.mx.Lock()

	if s.isHandled {
		s.mx.Unlock()
		return true, func() {}
	}

	return false, func() {
		s.isHandled = true
		s.mx.Unlock()
	}
}

func (s *handlesStatus) Check() bool {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.isHandled
}

func (s *handlesStatus) Set(v bool) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.isHandled = v
}
