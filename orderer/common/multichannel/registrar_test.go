/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package multichannel

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/osdi23p228/fabric/bccsp/sw"
	"github.com/osdi23p228/fabric/common/channelconfig"
	"github.com/osdi23p228/fabric/common/crypto/tlsgen"
	"github.com/osdi23p228/fabric/common/ledger/blockledger"
	"github.com/osdi23p228/fabric/common/ledger/blockledger/fileledger"
	"github.com/osdi23p228/fabric/common/metrics/disabled"
	"github.com/osdi23p228/fabric/common/policies"
	"github.com/osdi23p228/fabric/core/config/configtest"
	"github.com/osdi23p228/fabric/internal/configtxgen/encoder"
	"github.com/osdi23p228/fabric/internal/configtxgen/genesisconfig"
	"github.com/osdi23p228/fabric/internal/pkg/identity"
	"github.com/osdi23p228/fabric/orderer/common/blockcutter"
	"github.com/osdi23p228/fabric/orderer/common/localconfig"
	"github.com/osdi23p228/fabric/orderer/common/multichannel/mocks"
	"github.com/osdi23p228/fabric/orderer/common/types"
	"github.com/osdi23p228/fabric/orderer/consensus"
	"github.com/osdi23p228/fabric/protoutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:generate counterfeiter -o mocks/resources.go --fake-name Resources . resources

type resources interface {
	channelconfig.Resources
}

//go:generate counterfeiter -o mocks/orderer_config.go --fake-name OrdererConfig . ordererConfig

type ordererConfig interface {
	channelconfig.Orderer
}

//go:generate counterfeiter -o mocks/orderer_capabilities.go --fake-name OrdererCapabilities . ordererCapabilities

type ordererCapabilities interface {
	channelconfig.OrdererCapabilities
}

//go:generate counterfeiter -o mocks/channel_config.go --fake-name ChannelConfig . channelConfig

type channelConfig interface {
	channelconfig.Channel
}

//go:generate counterfeiter -o mocks/channel_capabilities.go --fake-name ChannelCapabilities . channelCapabilities

type channelCapabilities interface {
	channelconfig.ChannelCapabilities
}

//go:generate counterfeiter -o mocks/signer_serializer.go --fake-name SignerSerializer . signerSerializer

type signerSerializer interface {
	identity.SignerSerializer
}

func mockCrypto() *mocks.SignerSerializer {
	return &mocks.SignerSerializer{}
}

func newLedgerAndFactory(dir string, chainID string, genesisBlockSys *cb.Block) (blockledger.Factory, blockledger.ReadWriter) {
	rlf, err := fileledger.New(dir, &disabled.Provider{})
	if err != nil {
		panic(err)
	}

	rl, err := rlf.GetOrCreate(chainID)
	if err != nil {
		panic(err)
	}

	if genesisBlockSys != nil {
		err = rl.Append(genesisBlockSys)
		if err != nil {
			panic(err)
		}
	}
	return rlf, rl
}

func testMessageOrderAndRetrieval(maxMessageCount uint32, chainID string, chainSupport *ChainSupport, lr blockledger.ReadWriter, t *testing.T) {
	messages := make([]*cb.Envelope, maxMessageCount)
	for i := uint32(0); i < maxMessageCount; i++ {
		messages[i] = makeNormalTx(chainID, int(i))
	}
	for _, message := range messages {
		chainSupport.Order(message, 0)
	}
	it, _ := lr.Iterator(&ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 1}}})
	defer it.Close()
	block, status := it.Next()
	assert.Equal(t, cb.Status_SUCCESS, status, "Could not retrieve block")
	for i := uint32(0); i < maxMessageCount; i++ {
		assert.True(t, proto.Equal(messages[i], protoutil.ExtractEnvelopeOrPanic(block, int(i))), "Block contents wrong at index %d", i)
	}
}

func TestConfigTx(t *testing.T) {
	// system channel
	confSys := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	genesisBlockSys := encoder.New(confSys).GenesisBlock()

	// Tests for a normal channel which contains 3 config transactions and other
	// normal transactions to make sure the right one returned
	t.Run("GetConfigTx - ok", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		_, rl := newLedgerAndFactory(tmpdir, "testchannelid", genesisBlockSys)
		for i := 0; i < 5; i++ {
			rl.Append(blockledger.CreateNextBlock(rl, []*cb.Envelope{makeNormalTx("testchannelid", i)}))
		}
		rl.Append(blockledger.CreateNextBlock(rl, []*cb.Envelope{makeConfigTx("testchannelid", 5)}))
		ctx := makeConfigTx("testchannelid", 6)
		rl.Append(blockledger.CreateNextBlock(rl, []*cb.Envelope{ctx}))

		// block with LAST_CONFIG metadata in SIGNATURES field
		block := blockledger.CreateNextBlock(rl, []*cb.Envelope{makeNormalTx("testchannelid", 7)})
		blockSignatureValue := protoutil.MarshalOrPanic(&cb.OrdererBlockMetadata{
			LastConfig: &cb.LastConfig{Index: 7},
		})
		block.Metadata.Metadata[cb.BlockMetadataIndex_SIGNATURES] = protoutil.MarshalOrPanic(&cb.Metadata{Value: blockSignatureValue})
		rl.Append(block)

		pctx := configTx(rl)
		assert.True(t, proto.Equal(pctx, ctx), "Did not select most recent config transaction")
	})
}

func TestNewRegistrar(t *testing.T) {
	//system channel
	confSys := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	genesisBlockSys := encoder.New(confSys).GenesisBlock()

	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)

	// This test checks to make sure the orderer can come up if it cannot find any chains
	t.Run("No chains", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		lf, err := fileledger.New(tmpdir, &disabled.Provider{})
		require.NoError(t, err)

		consenters := make(map[string]consensus.Consenter)
		consenters["etcdraft"] = &mockConsenter{}

		var manager *Registrar
		assert.NotPanics(t, func() {
			manager = NewRegistrar(localconfig.TopLevel{}, lf, mockCrypto(), &disabled.Provider{}, cryptoProvider)
			manager.Initialize(consenters)
		}, "Should not panic when starting without a system channel")
		require.NotNil(t, manager)
		list := manager.ChannelList()
		assert.Equal(t, types.ChannelList{}, list)
		info, err := manager.ChannelInfo("my-channel")
		assert.EqualError(t, err, types.ErrChannelNotExist.Error())
		assert.Equal(t, types.ChannelInfo{}, info)
	})

	// This test checks to make sure that the orderer refuses to come up if there are multiple system channels
	t.Run("Multiple system chains - failure", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		lf, err := fileledger.New(tmpdir, &disabled.Provider{})
		require.NoError(t, err)

		for _, id := range []string{"foo", "bar"} {
			rl, err := lf.GetOrCreate(id)
			assert.NoError(t, err)

			err = rl.Append(encoder.New(confSys).GenesisBlockForChannel(id))
			assert.NoError(t, err)
		}

		consenters := make(map[string]consensus.Consenter)
		consenters[confSys.Orderer.OrdererType] = &mockConsenter{}

		assert.Panics(t, func() {
			NewRegistrar(localconfig.TopLevel{}, lf, mockCrypto(), &disabled.Provider{}, cryptoProvider).Initialize(consenters)
		}, "Two system channels should have caused panic")
	})

	// This test essentially brings the entire system up and is ultimately what main.go will replicate
	t.Run("Correct flow with system channel", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		lf, rl := newLedgerAndFactory(tmpdir, "testchannelid", genesisBlockSys)

		consenters := make(map[string]consensus.Consenter)
		consenters[confSys.Orderer.OrdererType] = &mockConsenter{}

		manager := NewRegistrar(localconfig.TopLevel{}, lf, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		manager.Initialize(consenters)

		chainSupport := manager.GetChain("Fake")
		assert.Nilf(t, chainSupport, "Should not have found a chain that was not created")

		chainSupport = manager.GetChain("testchannelid")
		assert.NotNilf(t, chainSupport, "Should have gotten chain which was initialized by ledger")

		list := manager.ChannelList()
		require.NotNil(t, list.SystemChannel)

		assert.Equal(
			t,
			types.ChannelList{
				SystemChannel: &types.ChannelInfoShort{Name: "testchannelid", URL: ""},
				Channels:      nil},
			list,
		)

		info, err := manager.ChannelInfo("testchannelid")
		assert.NoError(t, err)
		assert.Equal(t,
			types.ChannelInfo{Name: "testchannelid", URL: "", ClusterRelation: "none", Status: "active", Height: 1},
			info,
		)

		testMessageOrderAndRetrieval(confSys.Orderer.BatchSize.MaxMessageCount, "testchannelid", chainSupport, rl, t)
	})
}

func TestCreateChain(t *testing.T) {
	//system channel
	confSys := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	genesisBlockSys := encoder.New(confSys).GenesisBlock()

	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)

	t.Run("Create chain", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		lf, _ := newLedgerAndFactory(tmpdir, "testchannelid", genesisBlockSys)

		consenters := make(map[string]consensus.Consenter)
		consenters[confSys.Orderer.OrdererType] = &mockConsenter{cluster: true}

		manager := NewRegistrar(localconfig.TopLevel{}, lf, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		manager.Initialize(consenters)

		ledger, err := lf.GetOrCreate("mychannel")
		assert.NoError(t, err)

		genesisBlock := encoder.New(confSys).GenesisBlockForChannel("mychannel")
		ledger.Append(genesisBlock)

		// Before creating the chain, it doesn't exist
		assert.Nil(t, manager.GetChain("mychannel"))
		// After creating the chain, it exists
		manager.CreateChain("mychannel")
		chain := manager.GetChain("mychannel")
		assert.NotNil(t, chain)

		list := manager.ChannelList()
		assert.Equal(
			t,
			types.ChannelList{
				SystemChannel: &types.ChannelInfoShort{Name: "testchannelid", URL: ""},
				Channels:      []types.ChannelInfoShort{{Name: "mychannel", URL: ""}}},
			list,
		)

		info, err := manager.ChannelInfo("testchannelid")
		assert.NoError(t, err)
		assert.Equal(t,
			types.ChannelInfo{Name: "testchannelid", URL: "", ClusterRelation: types.ClusterRelationMember, Status: types.StatusActive, Height: 1},
			info,
		)

		info, err = manager.ChannelInfo("mychannel")
		assert.NoError(t, err)
		assert.Equal(t,
			types.ChannelInfo{Name: "mychannel", URL: "", ClusterRelation: types.ClusterRelationMember, Status: types.StatusActive, Height: 1},
			info,
		)

		// A subsequent creation, replaces the chain.
		manager.CreateChain("mychannel")
		chain2 := manager.GetChain("mychannel")
		assert.NotNil(t, chain2)
		// They are not the same
		assert.NotEqual(t, chain, chain2)
		// The old chain is halted
		_, ok := <-chain.Chain.(*mockChainCluster).queue
		assert.False(t, ok)

		// The new chain is not halted: Close the channel to prove that.
		close(chain2.Chain.(*mockChainCluster).queue)
	})

	// This test brings up the entire system, with the mock consenter, including the broadcasters etc. and creates a new chain
	t.Run("New chain", func(t *testing.T) {
		expectedLastConfigSeq := uint64(1)
		newChainID := "test-new-chain"

		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		lf, rl := newLedgerAndFactory(tmpdir, "testchannelid", genesisBlockSys)

		consenters := make(map[string]consensus.Consenter)
		consenters[confSys.Orderer.OrdererType] = &mockConsenter{}

		manager := NewRegistrar(localconfig.TopLevel{}, lf, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		manager.Initialize(consenters)
		orglessChannelConf := genesisconfig.Load(genesisconfig.SampleSingleMSPChannelProfile, configtest.GetDevConfigDir())
		orglessChannelConf.Application.Organizations = nil
		envConfigUpdate, err := encoder.MakeChannelCreationTransaction(newChainID, mockCrypto(), orglessChannelConf)
		assert.NoError(t, err, "Constructing chain creation tx")

		res, err := manager.NewChannelConfig(envConfigUpdate)
		assert.NoError(t, err, "Constructing initial channel config")

		configEnv, err := res.ConfigtxValidator().ProposeConfigUpdate(envConfigUpdate)
		assert.NoError(t, err, "Proposing initial update")
		assert.Equal(t, expectedLastConfigSeq, configEnv.GetConfig().Sequence, "Sequence of config envelope for new channel should always be set to %d", expectedLastConfigSeq)

		ingressTx, err := protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG, newChainID, mockCrypto(), configEnv, msgVersion, epoch)
		assert.NoError(t, err, "Creating ingresstx")

		wrapped := wrapConfigTx(ingressTx)

		chainSupport := manager.GetChain(manager.SystemChannelID())
		assert.NotNilf(t, chainSupport, "Could not find system channel")

		chainSupport.Configure(wrapped, 0)
		func() {
			it, _ := rl.Iterator(&ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 1}}})
			defer it.Close()
			block, status := it.Next()
			if status != cb.Status_SUCCESS {
				t.Fatalf("Could not retrieve block")
			}
			if len(block.Data.Data) != 1 {
				t.Fatalf("Should have had only one message in the orderer transaction block")
			}

			assert.True(t, proto.Equal(wrapped, protoutil.UnmarshalEnvelopeOrPanic(block.Data.Data[0])), "Orderer config block contains wrong transaction")
		}()

		chainSupport = manager.GetChain(newChainID)
		if chainSupport == nil {
			t.Fatalf("Should have gotten new chain which was created")
		}

		messages := make([]*cb.Envelope, confSys.Orderer.BatchSize.MaxMessageCount)
		for i := 0; i < int(confSys.Orderer.BatchSize.MaxMessageCount); i++ {
			messages[i] = makeNormalTx(newChainID, i)
		}

		for _, message := range messages {
			chainSupport.Order(message, 0)
		}

		it, _ := chainSupport.Reader().Iterator(&ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 0}}})
		defer it.Close()
		block, status := it.Next()
		if status != cb.Status_SUCCESS {
			t.Fatalf("Could not retrieve new chain genesis block")
		}
		if len(block.Data.Data) != 1 {
			t.Fatalf("Should have had only one message in the new genesis block")
		}

		assert.True(t, proto.Equal(ingressTx, protoutil.UnmarshalEnvelopeOrPanic(block.Data.Data[0])), "Genesis block contains wrong transaction")

		block, status = it.Next()
		if status != cb.Status_SUCCESS {
			t.Fatalf("Could not retrieve block on new chain")
		}
		for i := 0; i < int(confSys.Orderer.BatchSize.MaxMessageCount); i++ {
			if !proto.Equal(protoutil.ExtractEnvelopeOrPanic(block, i), messages[i]) {
				t.Errorf("Block contents wrong at index %d in new chain", i)
			}
		}

		cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
		assert.NoError(t, err)
		rcs, err := newChainSupport(manager, chainSupport.ledgerResources, consenters, mockCrypto(), blockcutter.NewMetrics(&disabled.Provider{}), cryptoProvider)
		assert.NoError(t, err)
		assert.Equal(t, expectedLastConfigSeq, rcs.lastConfigSeq, "On restart, incorrect lastConfigSeq")
	})
}

func TestResourcesCheck(t *testing.T) {
	mockOrderer := &mocks.OrdererConfig{}
	mockOrdererCaps := &mocks.OrdererCapabilities{}
	mockOrderer.CapabilitiesReturns(mockOrdererCaps)
	mockChannel := &mocks.ChannelConfig{}
	mockChannelCaps := &mocks.ChannelCapabilities{}
	mockChannel.CapabilitiesReturns(mockChannelCaps)

	mockResources := &mocks.Resources{}
	mockResources.PolicyManagerReturns(&policies.ManagerImpl{})

	t.Run("GoodResources", func(t *testing.T) {
		mockResources.OrdererConfigReturns(mockOrderer, true)
		mockResources.ChannelConfigReturns(mockChannel)

		err := checkResources(mockResources)
		assert.NoError(t, err)
	})

	t.Run("MissingOrdererConfigPanic", func(t *testing.T) {
		mockResources.OrdererConfigReturns(nil, false)

		err := checkResources(mockResources)
		assert.Error(t, err)
		assert.Regexp(t, "config does not contain orderer config", err.Error())
	})

	t.Run("MissingOrdererCapability", func(t *testing.T) {
		mockResources.OrdererConfigReturns(mockOrderer, true)
		mockOrdererCaps.SupportedReturns(errors.New("An error"))

		err := checkResources(mockResources)
		assert.Error(t, err)
		assert.Regexp(t, "config requires unsupported orderer capabilities:", err.Error())

		// reset
		mockOrdererCaps.SupportedReturns(nil)
	})

	t.Run("MissingChannelCapability", func(t *testing.T) {
		mockChannelCaps.SupportedReturns(errors.New("An error"))

		err := checkResources(mockResources)
		assert.Error(t, err)
		assert.Regexp(t, "config requires unsupported channel capabilities:", err.Error())
	})

	t.Run("MissingOrdererConfigPanic", func(t *testing.T) {
		mockResources.OrdererConfigReturns(nil, false)

		assert.Panics(t, func() {
			checkResourcesOrPanic(mockResources)
		})
	})
}

// The registrar's BroadcastChannelSupport implementation should reject message types which should not be processed directly.
func TestBroadcastChannelSupport(t *testing.T) {
	// system channel
	confSys := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	genesisBlockSys := encoder.New(confSys).GenesisBlock()

	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)

	t.Run("Rejection", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		ledgerFactory, _ := newLedgerAndFactory(tmpdir, "testchannelid", genesisBlockSys)
		mockConsenters := map[string]consensus.Consenter{confSys.Orderer.OrdererType: &mockConsenter{}}
		registrar := NewRegistrar(localconfig.TopLevel{}, ledgerFactory, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		registrar.Initialize(mockConsenters)
		randomValue := 1
		configTx := makeConfigTx("testchannelid", randomValue)
		_, _, _, err = registrar.BroadcastChannelSupport(configTx)
		assert.Error(t, err, "Messages of type HeaderType_CONFIG should return an error.")
	})

	t.Run("No system channel", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		ledgerFactory, _ := newLedgerAndFactory(tmpdir, "", nil)
		mockConsenters := map[string]consensus.Consenter{confSys.Orderer.OrdererType: &mockConsenter{}, "etcdraft": &mockConsenter{}}
		config := localconfig.TopLevel{}
		config.General.BootstrapMethod = "none"
		config.General.GenesisFile = ""
		registrar := NewRegistrar(config, ledgerFactory, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		registrar.Initialize(mockConsenters)
		configTx := makeConfigTxFull("testchannelid", 1)
		_, _, _, err = registrar.BroadcastChannelSupport(configTx)
		assert.Error(t, err)
		assert.Equal(t, "channel creation request not allowed because the orderer system channel is not defined", err.Error())
	})
}

func TestRegistrar_JoinChannel(t *testing.T) {
	// system channel
	confSys := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	genesisBlockSys := encoder.New(confSys).GenesisBlockForChannel("sys-channel")
	confApp := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	confApp.Consortiums = nil
	confApp.Consortium = ""
	genesisBlockApp := encoder.New(confApp).GenesisBlockForChannel("my-channel")

	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)

	t.Run("Reject join when system channel exists", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		ledgerFactory, _ := newLedgerAndFactory(tmpdir, "sys-channel", genesisBlockSys)
		mockConsenters := map[string]consensus.Consenter{confSys.Orderer.OrdererType: &mockConsenter{}}
		registrar := NewRegistrar(localconfig.TopLevel{}, ledgerFactory, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		registrar.Initialize(mockConsenters)

		info, err := registrar.JoinChannel("some-app-channel", &cb.Block{}, true)
		assert.EqualError(t, err, "system channel exists")
		assert.Equal(t, types.ChannelInfo{}, info)
	})

	t.Run("Reject join when channel exists", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		ledgerFactory, _ := newLedgerAndFactory(tmpdir, "", nil)
		mockConsenters := map[string]consensus.Consenter{confSys.Orderer.OrdererType: &mockConsenter{}, "etcdraft": &mockConsenter{}}
		config := localconfig.TopLevel{}
		config.General.BootstrapMethod = "none"
		config.General.GenesisFile = ""
		registrar := NewRegistrar(config, ledgerFactory, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		registrar.Initialize(mockConsenters)

		ledger, err := ledgerFactory.GetOrCreate("my-channel")
		assert.NoError(t, err)
		ledger.Append(genesisBlockApp)

		// Before creating the chain, it doesn't exist
		assert.Nil(t, registrar.GetChain("my-channel"))
		// After creating the chain, it exists
		registrar.CreateChain("my-channel")
		assert.NotNil(t, registrar.GetChain("my-channel"))

		info, err := registrar.JoinChannel("my-channel", &cb.Block{}, true)
		assert.EqualError(t, err, "channel already exists")
		assert.Equal(t, types.ChannelInfo{}, info)
	})

	t.Run("Reject system channel join when app channels exist", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		ledgerFactory, _ := newLedgerAndFactory(tmpdir, "", nil)
		mockConsenters := map[string]consensus.Consenter{confSys.Orderer.OrdererType: &mockConsenter{}, "etcdraft": &mockConsenter{}}
		config := localconfig.TopLevel{}
		config.General.BootstrapMethod = "none"
		config.General.GenesisFile = ""
		registrar := NewRegistrar(config, ledgerFactory, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		registrar.Initialize(mockConsenters)

		ledger, err := ledgerFactory.GetOrCreate("my-channel")
		assert.NoError(t, err)
		ledger.Append(genesisBlockApp)

		// Before creating the chain, it doesn't exist
		assert.Nil(t, registrar.GetChain("my-channel"))
		// After creating the chain, it exists
		registrar.CreateChain("my-channel")
		assert.NotNil(t, registrar.GetChain("my-channel"))

		info, err := registrar.JoinChannel("sys-channel", &cb.Block{}, false)
		assert.EqualError(t, err, "application channels already exist")
		assert.Equal(t, types.ChannelInfo{}, info)
	})

	t.Run("no etcdraft consenter without system channel", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		ledgerFactory, _ := newLedgerAndFactory(tmpdir, "", nil)
		mockConsenters := map[string]consensus.Consenter{"not-raft": &mockConsenter{}}

		config := localconfig.TopLevel{}
		config.General.BootstrapMethod = "none"
		config.General.GenesisFile = ""
		registrar := NewRegistrar(config, ledgerFactory, mockCrypto(), &disabled.Provider{}, cryptoProvider)

		assert.Panics(t, func() { registrar.Initialize(mockConsenters) })
	})

	t.Run("Join app channel as member without on boarding", func(t *testing.T) {
		tmpdir, err := ioutil.TempDir("", "registrar_test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpdir)

		tlsCA, _ := tlsgen.NewCA()

		confAppRaft := genesisconfig.Load(genesisconfig.SampleDevModeEtcdRaftProfile, configtest.GetDevConfigDir())
		confAppRaft.Consortiums = nil
		confAppRaft.Consortium = ""
		generateCertificates(t, confAppRaft, tlsCA, tmpdir)
		bootstrapper, err := encoder.NewBootstrapper(confAppRaft)
		require.NoError(t, err, "cannot create bootstrapper")
		genesisBlockAppRaft := bootstrapper.GenesisBlockForChannel("my-raft-channel")
		require.NotNil(t, genesisBlockAppRaft)

		ledgerFactory, _ := newLedgerAndFactory(tmpdir, "", nil)
		mockConsenters := map[string]consensus.Consenter{confAppRaft.Orderer.OrdererType: &mockConsenter{cluster: true}}
		config := localconfig.TopLevel{}
		config.General.BootstrapMethod = "none"
		config.General.GenesisFile = ""
		registrar := NewRegistrar(config, ledgerFactory, mockCrypto(), &disabled.Provider{}, cryptoProvider)
		registrar.Initialize(mockConsenters)

		// Before join the chain, it doesn't exist
		assert.Nil(t, registrar.GetChain("my-raft-channel"))

		info, err := registrar.JoinChannel("my-raft-channel", genesisBlockAppRaft, true)
		assert.NoError(t, err)
		assert.Equal(t, types.ChannelInfo{Name: "my-raft-channel", URL: "", ClusterRelation: "member", Status: "active", Height: 0x1}, info)
		// After creating the chain, it exists
		assert.NotNil(t, registrar.GetChain("my-raft-channel"))
	})
}

func generateCertificates(t *testing.T, confAppRaft *genesisconfig.Profile, tlsCA tlsgen.CA, certDir string) {
	for i, c := range confAppRaft.Orderer.EtcdRaft.Consenters {
		srvC, err := tlsCA.NewServerCertKeyPair(c.Host)
		require.NoError(t, err)
		srvP := path.Join(certDir, fmt.Sprintf("server%d.crt", i))
		err = ioutil.WriteFile(srvP, srvC.Cert, 0644)
		require.NoError(t, err)

		clnC, err := tlsCA.NewClientCertKeyPair()
		require.NoError(t, err)
		clnP := path.Join(certDir, fmt.Sprintf("client%d.crt", i))
		err = ioutil.WriteFile(clnP, clnC.Cert, 0644)
		require.NoError(t, err)

		c.ServerTlsCert = []byte(srvP)
		c.ClientTlsCert = []byte(clnP)
	}
}
