package manager

import "sync"

type MutexManager interface {
}

type mutexManager struct {
	mutexs map[string]Mutex
}

func NewMutexManager() MutexManager {
	m := &mutexManager{}
	return m
}

func (m *mutexManager) NewMutex(mutexName string) Mutex {
	if mu, ok := m.mutexs[mutexName]; ok {
		return mu
	}
	mu := &mutex{
		name:     mutexName,
		lock:     &sync.Mutex{},
		mutexLog: make([]string, 0),
	}
	m.mutexs[mutexName] = mu
	return mu
}

type Mutex interface {
	Lock(id string)
	Unlock(id string)
}

type mutex struct {
	name     string
	lock     *sync.Mutex
	mutexLog []string
}

func (mu *mutex) Lock(id string) {
	mu.lock.Lock()
	mu.mutexLog = append(mu.mutexLog, id)
}

func (mu *mutex) Unlock(id string) {
	mu.lock.Unlock()
}
