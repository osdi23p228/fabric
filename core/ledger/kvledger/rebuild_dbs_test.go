/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kvledger

import (
	"path/filepath"
	"testing"

	configtxtest "github.com/osdi23p228/fabric/common/configtx/test"
	"github.com/osdi23p228/fabric/common/ledger/util"
	"github.com/osdi23p228/fabric/core/ledger/mock"
	"github.com/stretchr/testify/require"
)

func TestRebuildDBs(t *testing.T) {
	conf, cleanup := testConfig(t)
	defer cleanup()
	provider := testutilNewProvider(conf, t, &mock.DeployedChaincodeInfoProvider{})

	numLedgers := 3
	for i := 0; i < numLedgers; i++ {
		genesisBlock, _ := configtxtest.MakeGenesisBlock(constructTestLedgerID(i))
		provider.Create(genesisBlock)
	}

	// rebuild should fail when provider is still open
	err := RebuildDBs(conf)
	require.Error(t, err, "as another peer node command is executing, wait for that command to complete its execution or terminate it before retrying")
	provider.Close()

	err = RebuildDBs(conf)
	require.NoError(t, err)

	// verify blockstoreIndex, configHistory, history, state, bookkeeper dbs are deleted
	rootFSPath := conf.RootFSPath
	empty, err := util.DirEmpty(filepath.Join(BlockStorePath(rootFSPath), "index"))
	require.NoError(t, err)
	require.True(t, empty)
	empty, err = util.DirEmpty(ConfigHistoryDBPath(rootFSPath))
	require.NoError(t, err)
	require.True(t, empty)
	empty, err = util.DirEmpty(HistoryDBPath(rootFSPath))
	require.NoError(t, err)
	require.True(t, empty)
	empty, err = util.DirEmpty(StateDBPath(rootFSPath))
	require.NoError(t, err)
	require.True(t, empty)
	empty, err = util.DirEmpty(BookkeeperDBPath(rootFSPath))
	require.NoError(t, err)
	require.True(t, empty)

	// rebuild again should be successful
	err = RebuildDBs(conf)
	require.NoError(t, err)
}
