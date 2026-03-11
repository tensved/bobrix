package ctx

import "sync"

type handlesStatus struct {
	isHandled bool
	mx        sync.Mutex

	done chan struct{}
	once sync.Once
}

func newHandlesStatus() *handlesStatus {
	return &handlesStatus{
		done: make(chan struct{}),
	}
}

// Done closes when handled becomes true (once)
func (s *handlesStatus) doneCh() <-chan struct{} {
	return s.done
}

func (s *handlesStatus) markDoneIfHandledLocked() {
	if s.isHandled {
		s.once.Do(func() { close(s.done) })
	}
}

func (s *handlesStatus) isHandledWithUnlocker() (bool, func()) {
	s.mx.Lock()
	handled := s.isHandled
	return handled, func() { s.mx.Unlock() }
}

func (s *handlesStatus) check() bool {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.isHandled
}

func (s *handlesStatus) set(v bool) {
	s.mx.Lock()
	s.isHandled = v
	s.markDoneIfHandledLocked()
	s.mx.Unlock()
}
