/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoutil_test

import (
	"testing"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/osdi23p228/fabric/protoutil"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigGroup(t *testing.T) {
	assert.Equal(t,
		&common.ConfigGroup{
			Groups:   make(map[string]*common.ConfigGroup),
			Values:   make(map[string]*common.ConfigValue),
			Policies: make(map[string]*common.ConfigPolicy),
		},
		protoutil.NewConfigGroup(),
	)
}
