package shutdown

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDependencyNode_Insert_NilNode(t *testing.T) {
	assert.NotPanics(t, func() {
		var nilNode *DependencyNode

		nilNode.Insert(gofakeit.City(), NewDependencyNode(gofakeit.City(), nil))
	})
}
