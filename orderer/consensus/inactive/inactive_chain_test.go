/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package inactive_test

import (
	"github.com/osdi23p228/fabric/orderer/common/types"
	"testing"

	"github.com/osdi23p228/fabric/orderer/consensus/inactive"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestInactiveChain(t *testing.T) {
	err := errors.New("foo")
	chain := &inactive.Chain{Err: err}

	assert.Equal(t, err, chain.Order(nil, 0))
	assert.Equal(t, err, chain.Configure(nil, 0))
	assert.Equal(t, err, chain.WaitReady())
	assert.NotPanics(t, chain.Start)
	assert.NotPanics(t, chain.Halt)
	_, open := <-chain.Errored()
	assert.False(t, open)

	cRel, status := chain.StatusReport()
	assert.Equal(t, types.ClusterRelationConfigTracker, cRel)
	assert.Equal(t, types.StatusInactive, status)
}
