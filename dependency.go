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

func (node *DependencyNode) Key() string {
	if node == nil {
		return ""
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	return node.key
}

func (node *DependencyNode) Nodes() []*DependencyNode {
	if node == nil {
		return nil
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	return node.children
}

func (node *DependencyNode) Find(key string) *DependencyNode {
	if node == nil {
		return nil
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.key == key {
		return node
	}

	for _, child := range node.children {
		if found := child.Find(key); found != nil {
			return found
		}
	}

	return nil
}

func (node *DependencyNode) Insert(rootKey string, newNode *DependencyNode) {
	if node == nil || newNode == nil {
		return
	}

	if node.key == rootKey {
		node.mu.Lock()
		node.children = append(node.children, newNode)
		node.mu.Unlock()

		return
	}

	for _, child := range node.children {
		child.Insert(rootKey, newNode)
	}
}

func (node *DependencyNode) Shutdown(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if node == nil {
		return
	}

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.callback != nil {
		node.callback(ctx)
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(node.children))

	for _, child := range node.children {
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
