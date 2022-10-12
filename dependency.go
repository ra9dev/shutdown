package shutdown

import (
	"context"
	"fmt"
	"sync"
)

const dependenciesRootKey = "shutdown_dependencies_root"

type DependencyNode struct {
	key      string
	callback CallbackFunc
	children []*DependencyNode
	mu       *sync.RWMutex
}

func NewDependencyNode(key string, callback CallbackFunc) *DependencyNode {
	return &DependencyNode{
		key:      key,
		callback: callback,
		children: make([]*DependencyNode, 0),
		mu:       new(sync.RWMutex),
	}
}

func (n *DependencyNode) Key() string {
	if n == nil {
		return ""
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.key
}

func (n *DependencyNode) Nodes() []*DependencyNode {
	if n == nil {
		return nil
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.children
}

func (n *DependencyNode) ResetNodes() {
	if n == nil {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	n.children = nil
}

func (n *DependencyNode) HasNode(key string) bool {
	if n == nil {
		return false
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, child := range n.children {
		if child.key == key {
			return true
		}
	}

	return false
}

func (n *DependencyNode) Find(key string) *DependencyNode {
	if n == nil {
		return nil
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.key == key {
		return n
	}

	for _, child := range n.children {
		if found := child.Find(key); found != nil {
			return found
		}
	}

	return nil
}

func (n *DependencyNode) Insert(rootKey string, newNode *DependencyNode) {
	if n == nil || newNode == nil {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.key == rootKey {
		n.children = append(n.children, newNode)

		return
	}

	for _, child := range n.children {
		child.Insert(rootKey, newNode)
	}
}

func (n *DependencyNode) Shutdown(ctx context.Context) {
	if n == nil {
		return
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	wg := new(sync.WaitGroup)
	wg.Add(len(n.children))

	for _, child := range n.children {
		threadSafeChild := child

		go func() {
			defer wg.Done()

			threadSafeChild.Shutdown(ctx)
		}()
	}

	wg.Wait()

	if n.callback != nil {
		n.callback(ctx)
	}
}

type DependencyGraph struct {
	root *DependencyNode
}

func NewDependencyGraph() DependencyGraph {
	return DependencyGraph{
		root: NewDependencyNode(dependenciesRootKey, nil),
	}
}

func (dg DependencyGraph) Insert(rootKey string, newNode *DependencyNode) error {
	if newKeyInTree := dg.root.Find(newNode.key); newKeyInTree != nil && newKeyInTree.HasNode(rootKey) {
		return fmt.Errorf("%w: %s <-> %s", ErrCyclicDependencies, rootKey, newNode.key)
	}

	dg.root.Insert(rootKey, newNode)

	return nil
}

func (dg DependencyGraph) Shutdown(ctx context.Context) {
	dg.root.Shutdown(ctx)
}
