package collyresponsible

import "sync"

type VisitMap struct {
	visited map[string]bool
	lock    *sync.RWMutex
}

func (v *VisitMap) Add(url string) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.visited[url] = true
}

func (v *VisitMap) IsVisited(url string) bool {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.visited[url]
}

func NewVisitMap() *VisitMap {
	return &VisitMap{
		visited: make(map[string]bool),
		lock:    &sync.RWMutex{},
	}
}
