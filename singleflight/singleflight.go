// Package singleflight solves the problem of thundering herd.
package singleflight

import "sync"

// doCall is an in-flight or completed call to Do func.
type doCall struct {
	wg  sync.WaitGroup
	val any
	err error
}

// Group represents a class of work and forms a namespace in which units
// of work can be executed with duplicate suppression.
type Group struct {
	mu      sync.Mutex         // protects callMap
	callMap map[string]*doCall // lazily initialized
}

// Do executes and returns the results of the given function, making sure
// that only one execution is in-flight for a given key at a time. If a
// duplicate request comes in, the duplicate caller waits for the original
// to complete and receives the same results.
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.callMap == nil {
		g.callMap = make(map[string]*doCall)
	}
	if call, ok := g.callMap[key]; ok {
		g.mu.Unlock()
		call.wg.Wait()
		return call.val, call.err
	}
	call := new(doCall)
	call.wg.Add(1)
	g.callMap[key] = call
	g.mu.Unlock()
	call.val, call.err = fn()
	call.wg.Done()

	g.mu.Lock()
	delete(g.callMap, key)
	g.mu.Unlock()
	return call.val, call.err
}
