package operator

import (
	"testing"

	"github.com/seeker-insurance/kit/random"
	"github.com/stretchr/testify/assert"
)

func TestNeg(t *testing.T) {
	A, _ := random.Int64s(25)
	for i := range A {
		a := int(A[i])
		assert.Equal(t, -a, Neg(a))
	}
}

func TestNoOp(t *testing.T) {
	A, _ := random.Int64s(25)
	for i := range A {
		a := int(A[i])
		assert.Equal(t, a, NoOp(a))
	}
}
