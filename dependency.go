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

	if n.callback != nil {
		n.callback(ctx)
	}

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
}

type DependencyTree struct {
	root *DependencyNode
}

func NewDependencyTree() DependencyTree {
	return DependencyTree{
		root: NewDependencyNode(dependenciesRootKey, nil),
	}
}

func (dg DependencyTree) Insert(rootKey string, newNode *DependencyNode) error {
	if newKeyInTree := dg.root.Find(newNode.key); newKeyInTree != nil {
		if rootNodeInNew := newKeyInTree.Find(rootKey); rootNodeInNew != nil {
			return fmt.Errorf("%w: %s <-> %s", ErrCyclicDependencies, rootKey, newNode.key)
		}
	}

	if rootKeyInTree := dg.root.Find(rootKey); rootKeyInTree == nil {
		return fmt.Errorf("%w for key %s", ErrNoDependencyRoot, rootKey)
	}

	dg.root.Insert(rootKey, newNode)

	return nil
}

func (dg DependencyTree) Shutdown(ctx context.Context) {
	dg.root.Shutdown(ctx)
}
