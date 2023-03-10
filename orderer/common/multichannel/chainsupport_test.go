/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package multichannel

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/osdi23p228/fabric/bccsp/sw"
	"github.com/osdi23p228/fabric/common/channelconfig"
	"github.com/osdi23p228/fabric/common/deliver/mock"
	"github.com/osdi23p228/fabric/common/ledger/blockledger"
	"github.com/osdi23p228/fabric/common/policies"
	"github.com/osdi23p228/fabric/core/config/configtest"
	"github.com/osdi23p228/fabric/internal/configtxgen/encoder"
	"github.com/osdi23p228/fabric/internal/configtxgen/genesisconfig"
	msgprocessormocks "github.com/osdi23p228/fabric/orderer/common/msgprocessor/mocks"
	"github.com/osdi23p228/fabric/orderer/common/multichannel/mocks"
	"github.com/osdi23p228/fabric/protoutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//go:generate counterfeiter -o mocks/policy.go --fake-name Policy . policy
//go:generate counterfeiter -o mocks/policy_manager.go --fake-name PolicyManager . policyManager
//go:generate counterfeiter -o mocks/read_writer.go --fake-name ReadWriter . readWriter

type policy interface{ policies.Policy }
type policyManager interface{ policies.Manager }
type readWriter interface{ blockledger.ReadWriter }

func TestChainSupportBlock(t *testing.T) {
	ledger := &mocks.ReadWriter{}
	ledger.HeightReturns(100)
	iterator := &mock.BlockIterator{}
	iterator.NextReturns(&common.Block{Header: &common.BlockHeader{Number: 99}}, common.Status_SUCCESS)
	ledger.IteratorReturns(iterator, 99)
	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)
	cs := &ChainSupport{
		ledgerResources: &ledgerResources{ReadWriter: ledger},
		BCCSP:           cryptoProvider,
	}

	assert.Nil(t, cs.Block(100))
	assert.Equal(t, uint64(99), cs.Block(99).Header.Number)
}

type mutableResourcesMock struct {
	*mocks.Resources
	newConsensusMetadataVal []byte
}

func (*mutableResourcesMock) Update(*channelconfig.Bundle) {
	panic("implement me")
}

func (mrm *mutableResourcesMock) CreateBundle(channelID string, c *common.Config) (channelconfig.Resources, error) {
	mockOrderer := &mocks.OrdererConfig{}
	mockOrderer.ConsensusMetadataReturns(mrm.newConsensusMetadataVal)
	mockResources := &mocks.Resources{}
	mockResources.OrdererConfigReturns(mockOrderer, true)

	return mockResources, nil

}

func TestVerifyBlockSignature(t *testing.T) {
	mockResources := &mocks.Resources{}
	mockValidator := &mocks.ConfigTXValidator{}
	mockValidator.ChannelIDReturns("mychannel")
	mockResources.ConfigtxValidatorReturns(mockValidator)

	mockPolicy := &mocks.Policy{}
	mockPolicyManager := &mocks.PolicyManager{}
	mockResources.PolicyManagerReturns(mockPolicyManager)

	ms := &mutableResourcesMock{
		Resources: mockResources,
	}
	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)
	cs := &ChainSupport{
		ledgerResources: &ledgerResources{
			configResources: &configResources{
				mutableResources: ms,
				bccsp:            cryptoProvider,
			},
		},
		BCCSP: cryptoProvider,
	}

	// Scenario I: Policy manager isn't initialized
	// and thus policy cannot be found
	mockPolicyManager.GetPolicyReturns(nil, false)
	err = cs.VerifyBlockSignature([]*protoutil.SignedData{}, nil)
	assert.EqualError(t, err, "policy /Channel/Orderer/BlockValidation wasn't found")

	mockPolicyManager.GetPolicyReturns(mockPolicy, true)
	// Scenario II: Policy manager finds policy, but it evaluates
	// to error.
	mockPolicy.EvaluateSignedDataReturns(errors.New("invalid signature"))
	err = cs.VerifyBlockSignature([]*protoutil.SignedData{}, nil)
	assert.EqualError(t, err, "block verification failed: invalid signature")

	// Scenario III: Policy manager finds policy, and it evaluates to success
	mockPolicy.EvaluateSignedDataReturns(nil)
	assert.NoError(t, cs.VerifyBlockSignature([]*protoutil.SignedData{}, nil))

	// Scenario IV: A bad config envelope is passed
	err = cs.VerifyBlockSignature([]*protoutil.SignedData{}, &common.ConfigEnvelope{})
	assert.EqualError(t, err, "channelconfig Config cannot be nil")

	// Scenario V: A valid config envelope is passed
	assert.NoError(t, cs.VerifyBlockSignature([]*protoutil.SignedData{}, testConfigEnvelope(t)))

}

func TestConsensusMetadataValidation(t *testing.T) {
	oldConsensusMetadata := []byte("old consensus metadata")
	newConsensusMetadata := []byte("new consensus metadata")
	mockValidator := &mocks.ConfigTXValidator{}
	mockValidator.ChannelIDReturns("mychannel")
	mockValidator.ProposeConfigUpdateReturns(testConfigEnvelope(t), nil)
	mockOrderer := &mocks.OrdererConfig{}
	mockOrderer.ConsensusMetadataReturns(oldConsensusMetadata)
	mockResources := &mocks.Resources{}
	mockResources.ConfigtxValidatorReturns(mockValidator)
	mockResources.OrdererConfigReturns(mockOrderer, true)

	ms := &mutableResourcesMock{
		Resources:               mockResources,
		newConsensusMetadataVal: newConsensusMetadata,
	}
	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)
	mv := &msgprocessormocks.MetadataValidator{}
	cs := &ChainSupport{
		ledgerResources: &ledgerResources{
			configResources: &configResources{
				mutableResources: ms,
				bccsp:            cryptoProvider,
			},
		},
		MetadataValidator: mv,
		BCCSP:             cryptoProvider,
	}

	// case 1: valid consensus metadata update
	_, err = cs.ProposeConfigUpdate(&common.Envelope{})
	assert.NoError(t, err)

	// validate arguments to ValidateConsensusMetadata
	assert.Equal(t, 1, mv.ValidateConsensusMetadataCallCount())
	om, nm, nc := mv.ValidateConsensusMetadataArgsForCall(0)
	assert.False(t, nc)
	assert.Equal(t, oldConsensusMetadata, om.ConsensusMetadata())
	assert.Equal(t, newConsensusMetadata, nm.ConsensusMetadata())

	// case 2: invalid consensus metadata update
	mv.ValidateConsensusMetadataReturns(errors.New("bananas"))
	_, err = cs.ProposeConfigUpdate(&common.Envelope{})
	assert.EqualError(t, err, "consensus metadata update for channel config update is invalid: bananas")
}

func testConfigEnvelope(t *testing.T) *common.ConfigEnvelope {
	conf := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	group, err := encoder.NewChannelGroup(conf)
	assert.NoError(t, err)
	group.Groups["Orderer"].Values["ConsensusType"].Value, err = proto.Marshal(&orderer.ConsensusType{
		Metadata: []byte("new consensus metadata"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, group)
	return &common.ConfigEnvelope{
		Config: &common.Config{
			ChannelGroup: group,
		},
	}
}
