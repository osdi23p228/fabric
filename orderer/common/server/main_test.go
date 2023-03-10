// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/osdi23p228/fabric/bccsp/factory"
	"github.com/osdi23p228/fabric/bccsp/sw"
	"github.com/osdi23p228/fabric/common/channelconfig"
	"github.com/osdi23p228/fabric/common/crypto/tlsgen"
	"github.com/osdi23p228/fabric/common/flogging"
	"github.com/osdi23p228/fabric/common/flogging/floggingtest"
	"github.com/osdi23p228/fabric/common/ledger/blockledger"
	"github.com/osdi23p228/fabric/common/ledger/blockledger/fileledger"
	"github.com/osdi23p228/fabric/common/metrics/disabled"
	"github.com/osdi23p228/fabric/common/metrics/prometheus"
	"github.com/osdi23p228/fabric/core/config/configtest"
	"github.com/osdi23p228/fabric/internal/configtxgen/encoder"
	"github.com/osdi23p228/fabric/internal/configtxgen/genesisconfig"
	"github.com/osdi23p228/fabric/internal/pkg/comm"
	"github.com/osdi23p228/fabric/internal/pkg/identity"
	"github.com/osdi23p228/fabric/orderer/common/bootstrap/file"
	"github.com/osdi23p228/fabric/orderer/common/cluster"
	"github.com/osdi23p228/fabric/orderer/common/localconfig"
	"github.com/osdi23p228/fabric/orderer/common/multichannel"
	"github.com/osdi23p228/fabric/orderer/common/onboarding"
	server_mocks "github.com/osdi23p228/fabric/orderer/common/server/mocks"
	"github.com/osdi23p228/fabric/orderer/consensus"
	"github.com/osdi23p228/fabric/protoutil"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate counterfeiter -o mocks/signer_serializer.go --fake-name SignerSerializer . signerSerializer

type signerSerializer interface {
	identity.SignerSerializer
}

// the path to cryptogen, which can be used by tests to create certificates
var cryptogen, tempDir string

func TestMain(m *testing.M) {
	var err error
	cryptogen, err = gexec.Build("github.com/osdi23p228/fabric/cmd/cryptogen")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cryptogen build failed: %v", err)
		os.Exit(-1)
	}
	defer gexec.CleanupBuildArtifacts()

	tempDir, err = ioutil.TempDir("", "main-test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temporary directory: %v", err)
		os.Exit(-1)
	}
	defer os.RemoveAll(tempDir)

	copyYamlFiles("testdata", tempDir)

	os.Exit(m.Run())
}

func copyYamlFiles(src, dst string) {
	for _, file := range []string{"configtx.yaml", "examplecom-config.yaml", "orderer.yaml"} {
		fileBytes, err := ioutil.ReadFile(filepath.Join(src, file))
		if err != nil {
			os.Exit(-1)
		}
		err = ioutil.WriteFile(filepath.Join(dst, file), fileBytes, 0644)
		if err != nil {
			os.Exit(-1)
		}
	}
}

func TestInitializeLogging(t *testing.T) {
	origEnvValue := os.Getenv("FABRIC_LOGGING_SPEC")
	os.Setenv("FABRIC_LOGGING_SPEC", "foo=debug")
	initializeLogging()
	assert.Equal(t, "debug", flogging.LoggerLevel("foo"))
	os.Setenv("FABRIC_LOGGING_SPEC", origEnvValue)
}

func TestInitializeProfilingService(t *testing.T) {
	origEnvValue := os.Getenv("FABRIC_LOGGING_SPEC")
	defer os.Setenv("FABRIC_LOGGING_SPEC", origEnvValue)
	os.Setenv("FABRIC_LOGGING_SPEC", "debug")
	// get a free random port
	listenAddr := func() string {
		l, _ := net.Listen("tcp", "localhost:0")
		l.Close()
		return l.Addr().String()
	}()
	go initializeProfilingService(
		&localconfig.TopLevel{
			General: localconfig.General{
				Profile: localconfig.Profile{
					Enabled: true,
					Address: listenAddr,
				}},
			Kafka: localconfig.Kafka{Verbose: true},
		},
	)
	time.Sleep(500 * time.Millisecond)
	if _, err := http.Get("http://" + listenAddr + "/" + "/debug/"); err != nil {
		t.Logf("Expected pprof to be up (will retry again in 3 seconds): %s", err)
		time.Sleep(3 * time.Second)
		if _, err := http.Get("http://" + listenAddr + "/" + "/debug/"); err != nil {
			t.Fatalf("Expected pprof to be up: %s", err)
		}
	}
}

func TestInitializeServerConfig(t *testing.T) {
	conf := &localconfig.TopLevel{
		General: localconfig.General{
			ConnectionTimeout: 7 * time.Second,
			TLS: localconfig.TLS{
				Enabled:            true,
				ClientAuthRequired: true,
				Certificate:        "main.go",
				PrivateKey:         "main.go",
				RootCAs:            []string{"main.go"},
				ClientRootCAs:      []string{"main.go"},
			},
		},
	}
	sc := initializeServerConfig(conf, nil)
	expectedContent, _ := ioutil.ReadFile("main.go")
	assert.Equal(t, expectedContent, sc.SecOpts.Certificate)
	assert.Equal(t, expectedContent, sc.SecOpts.Key)
	assert.Equal(t, [][]byte{expectedContent}, sc.SecOpts.ServerRootCAs)
	assert.Equal(t, [][]byte{expectedContent}, sc.SecOpts.ClientRootCAs)

	sc = initializeServerConfig(conf, nil)
	defaultOpts := comm.DefaultKeepaliveOptions
	assert.Equal(t, defaultOpts.ServerMinInterval, sc.KaOpts.ServerMinInterval)
	assert.Equal(t, time.Duration(0), sc.KaOpts.ServerInterval)
	assert.Equal(t, time.Duration(0), sc.KaOpts.ServerTimeout)
	assert.Equal(t, 7*time.Second, sc.ConnectionTimeout)
	testDuration := 10 * time.Second
	conf.General.Keepalive = localconfig.Keepalive{
		ServerMinInterval: testDuration,
		ServerInterval:    testDuration,
		ServerTimeout:     testDuration,
	}
	sc = initializeServerConfig(conf, nil)
	assert.Equal(t, testDuration, sc.KaOpts.ServerMinInterval)
	assert.Equal(t, testDuration, sc.KaOpts.ServerInterval)
	assert.Equal(t, testDuration, sc.KaOpts.ServerTimeout)

	sc = initializeServerConfig(conf, nil)
	assert.NotNil(t, sc.Logger)
	assert.Equal(t, comm.NewServerStatsHandler(&disabled.Provider{}), sc.ServerStatsHandler)
	assert.Len(t, sc.UnaryInterceptors, 2)
	assert.Len(t, sc.StreamInterceptors, 2)

	sc = initializeServerConfig(conf, &prometheus.Provider{})
	assert.NotNil(t, sc.ServerStatsHandler)

	goodFile := "main.go"
	badFile := "does_not_exist"

	oldLogger := logger
	defer func() { logger = oldLogger }()
	logger, _ = floggingtest.NewTestLogger(t)

	testCases := []struct {
		name           string
		certificate    string
		privateKey     string
		rootCA         string
		clientRootCert string
		clusterCert    string
		clusterKey     string
		clusterCA      string
		isCluster      bool
	}{
		{
			name:           "BadCertificate",
			certificate:    badFile,
			privateKey:     goodFile,
			rootCA:         goodFile,
			clientRootCert: goodFile,
		},
		{
			name:           "BadPrivateKey",
			certificate:    goodFile,
			privateKey:     badFile,
			rootCA:         goodFile,
			clientRootCert: goodFile,
		},
		{
			name:           "BadRootCA",
			certificate:    goodFile,
			privateKey:     goodFile,
			rootCA:         badFile,
			clientRootCert: goodFile,
		},
		{
			name:           "BadClientRootCertificate",
			certificate:    goodFile,
			privateKey:     goodFile,
			rootCA:         goodFile,
			clientRootCert: badFile,
		},
		{
			name:           "BadCertificate - cluster reuses server config",
			certificate:    badFile,
			privateKey:     goodFile,
			rootCA:         goodFile,
			clientRootCert: goodFile,
			clusterCert:    "",
			clusterKey:     "",
			clusterCA:      "",
			isCluster:      true,
		},
		{
			name:           "BadPrivateKey - cluster reuses server config",
			certificate:    goodFile,
			privateKey:     badFile,
			rootCA:         goodFile,
			clientRootCert: goodFile,
			clusterCert:    "",
			clusterKey:     "",
			clusterCA:      "",
			isCluster:      true,
		},
		{
			name:           "BadRootCA - cluster reuses server config",
			certificate:    goodFile,
			privateKey:     goodFile,
			rootCA:         badFile,
			clientRootCert: goodFile,
			clusterCert:    "",
			clusterKey:     "",
			clusterCA:      "",
			isCluster:      true,
		},
		{
			name:           "ClusterBadCertificate",
			certificate:    goodFile,
			privateKey:     goodFile,
			rootCA:         goodFile,
			clientRootCert: goodFile,
			clusterCert:    badFile,
			clusterKey:     goodFile,
			clusterCA:      goodFile,
			isCluster:      true,
		},
		{
			name:           "ClusterBadPrivateKey",
			certificate:    goodFile,
			privateKey:     goodFile,
			rootCA:         goodFile,
			clientRootCert: goodFile,
			clusterCert:    goodFile,
			clusterKey:     badFile,
			clusterCA:      goodFile,
			isCluster:      true,
		},
		{
			name:           "ClusterBadRootCA",
			certificate:    goodFile,
			privateKey:     goodFile,
			rootCA:         goodFile,
			clientRootCert: goodFile,
			clusterCert:    goodFile,
			clusterKey:     goodFile,
			clusterCA:      badFile,
			isCluster:      true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf := &localconfig.TopLevel{
				General: localconfig.General{
					TLS: localconfig.TLS{
						Enabled:            true,
						ClientAuthRequired: true,
						Certificate:        tc.certificate,
						PrivateKey:         tc.privateKey,
						RootCAs:            []string{tc.rootCA},
						ClientRootCAs:      []string{tc.clientRootCert},
					},
					Cluster: localconfig.Cluster{
						ClientCertificate: tc.clusterCert,
						ClientPrivateKey:  tc.clusterKey,
						RootCAs:           []string{tc.clusterCA},
					},
				},
			}
			require.Panics(t, func() {
				if !tc.isCluster {
					initializeServerConfig(conf, nil)
				} else {
					initializeClusterClientConfig(conf)
				}
			},
			)
		})
	}
}

func TestInitializeBootstrapChannel(t *testing.T) {
	cleanup := configtest.SetDevFabricConfigPath(t)
	defer cleanup()

	genesisFile := produceGenesisFile(t, genesisconfig.SampleSingleMSPSoloProfile, "testchannelid")
	defer os.Remove(genesisFile)

	fileLedgerLocation, _ := ioutil.TempDir("", "main_test-")
	defer os.RemoveAll(fileLedgerLocation)

	ledgerFactory, _, err := createLedgerFactory(
		&localconfig.TopLevel{
			FileLedger: localconfig.FileLedger{
				Location: fileLedgerLocation,
			},
		},
		&disabled.Provider{},
	)
	assert.NoError(t, err)
	bootstrapConfig := &localconfig.TopLevel{
		General: localconfig.General{
			BootstrapMethod: "file",
			BootstrapFile:   genesisFile,
		},
	}

	bootstrapBlock := extractBootstrapBlock(bootstrapConfig)
	initializeBootstrapChannel(bootstrapBlock, ledgerFactory)

	ledger, err := ledgerFactory.GetOrCreate("testchannelid")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), ledger.Height())
}

func TestExtractBootstrapBlock(t *testing.T) {
	cleanup := configtest.SetDevFabricConfigPath(t)
	defer cleanup()

	genesisFile := produceGenesisFile(t, genesisconfig.SampleSingleMSPSoloProfile, "testchannelid")
	defer os.Remove(genesisFile)

	tests := []struct {
		config *localconfig.TopLevel
		block  *common.Block
	}{
		{
			config: &localconfig.TopLevel{
				General: localconfig.General{BootstrapMethod: "file", BootstrapFile: genesisFile},
			},
			block: file.New(genesisFile).GenesisBlock(),
		},
		{
			config: &localconfig.TopLevel{
				General: localconfig.General{BootstrapMethod: "none"},
			},
			block: nil,
		},
	}
	for _, tt := range tests {
		b := extractBootstrapBlock(tt.config)
		assert.Truef(t, proto.Equal(tt.block, b), "wanted %v, got %v", tt.block, b)
	}
}

func TestExtractSysChanLastConfig(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "main_test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	rlf, err := fileledger.New(tmpdir, &disabled.Provider{})
	require.NoError(t, err)

	conf := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	genesisBlock := encoder.New(conf).GenesisBlock()

	lastConf := extractSysChanLastConfig(rlf, genesisBlock)
	assert.Nil(t, lastConf)

	rl, err := rlf.GetOrCreate("testchannelid")
	require.NoError(t, err)

	err = rl.Append(genesisBlock)
	require.NoError(t, err)

	lastConf = extractSysChanLastConfig(rlf, genesisBlock)
	assert.NotNil(t, lastConf)
	assert.Equal(t, uint64(0), lastConf.Header.Number)

	assert.Panics(t, func() {
		_ = extractSysChanLastConfig(rlf, nil)
	})

	configTx, err := protoutil.CreateSignedEnvelope(common.HeaderType_CONFIG, "testchannelid", nil, &common.ConfigEnvelope{}, 0, 0)
	require.NoError(t, err)

	nextBlock := blockledger.CreateNextBlock(rl, []*common.Envelope{configTx})
	nextBlock.Metadata.Metadata[common.BlockMetadataIndex_SIGNATURES] = protoutil.MarshalOrPanic(&common.Metadata{
		Value: protoutil.MarshalOrPanic(&common.OrdererBlockMetadata{
			LastConfig: &common.LastConfig{Index: rl.Height()},
		}),
	})
	nextBlock.Metadata.Metadata[common.BlockMetadataIndex_LAST_CONFIG] = protoutil.MarshalOrPanic(&common.Metadata{
		Value: protoutil.MarshalOrPanic(&common.LastConfig{Index: rl.Height()}),
	})
	err = rl.Append(nextBlock)
	require.NoError(t, err)

	lastConf = extractSysChanLastConfig(rlf, genesisBlock)
	assert.NotNil(t, lastConf)
	assert.Equal(t, uint64(1), lastConf.Header.Number)
}

func TestSelectClusterBootBlock(t *testing.T) {
	bootstrapBlock := &common.Block{Header: &common.BlockHeader{Number: 100}}
	lastConfBlock := &common.Block{Header: &common.BlockHeader{Number: 100}}

	clusterBoot := selectClusterBootBlock(bootstrapBlock, nil)
	assert.NotNil(t, clusterBoot)
	assert.Equal(t, uint64(100), clusterBoot.Header.Number)
	assert.True(t, bootstrapBlock == clusterBoot)

	clusterBoot = selectClusterBootBlock(bootstrapBlock, lastConfBlock)
	assert.NotNil(t, clusterBoot)
	assert.Equal(t, uint64(100), clusterBoot.Header.Number)
	assert.True(t, bootstrapBlock == clusterBoot)

	lastConfBlock.Header.Number = 200
	clusterBoot = selectClusterBootBlock(bootstrapBlock, lastConfBlock)
	assert.NotNil(t, clusterBoot)
	assert.Equal(t, uint64(200), clusterBoot.Header.Number)
	assert.True(t, lastConfBlock == clusterBoot)

	bootstrapBlock.Header.Number = 300
	clusterBoot = selectClusterBootBlock(bootstrapBlock, lastConfBlock)
	assert.NotNil(t, clusterBoot)
	assert.Equal(t, uint64(300), clusterBoot.Header.Number)
	assert.True(t, bootstrapBlock == clusterBoot)
}

func TestLoadLocalMSP(t *testing.T) {
	t.Run("Happy", func(t *testing.T) {
		localMSPDir := configtest.GetDevMspDir()
		localMSP := loadLocalMSP(
			&localconfig.TopLevel{
				General: localconfig.General{
					LocalMSPDir: localMSPDir,
					LocalMSPID:  "SampleOrg",
					BCCSP: &factory.FactoryOpts{
						ProviderName: "SW",
						SwOpts: &factory.SwOpts{
							HashFamily: "SHA2",
							SecLevel:   256,
						},
					},
				},
			},
		)
		require.NotNil(t, localMSP)
		id, err := localMSP.GetIdentifier()
		require.NoError(t, err)
		require.Equal(t, id, "SampleOrg")
	})

	t.Run("Error", func(t *testing.T) {
		oldLogger := logger
		defer func() { logger = oldLogger }()
		logger, _ = floggingtest.NewTestLogger(t)

		assert.Panics(t, func() {
			loadLocalMSP(
				&localconfig.TopLevel{
					General: localconfig.General{
						LocalMSPDir: "",
						LocalMSPID:  "",
					},
				},
			)
		})
	})
}

func TestInitializeMultichannelRegistrar(t *testing.T) {
	cleanup := configtest.SetDevFabricConfigPath(t)
	defer cleanup()
	genesisFile := produceGenesisFile(t, genesisconfig.SampleDevModeSoloProfile, "testchannelid")
	defer os.Remove(genesisFile)

	conf := genesisConfig(t, genesisFile)
	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)

	signer := &server_mocks.SignerSerializer{}

	t.Run("registrar with a system channel", func(t *testing.T) {
		lf, _, err := createLedgerFactory(conf, &disabled.Provider{})
		assert.NoError(t, err)
		bootBlock := file.New(genesisFile).GenesisBlock()
		initializeBootstrapChannel(bootBlock, lf)
		registrar := initializeMultichannelRegistrar(
			bootBlock,
			onboarding.NewReplicationInitiator(lf, bootBlock, conf, comm.SecureOptions{}, signer, cryptoProvider),
			&cluster.PredicateDialer{},
			comm.ServerConfig{},
			nil,
			conf,
			signer,
			&disabled.Provider{},
			&server_mocks.HealthChecker{},
			lf,
			cryptoProvider,
		)
		assert.NotNil(t, registrar)
		assert.Equal(t, "testchannelid", registrar.SystemChannelID())
	})

	t.Run("registrar without a system channel", func(t *testing.T) {
		conf.General.BootstrapMethod = "none"
		conf.General.GenesisFile = ""
		srv, err := comm.NewGRPCServer("127.0.0.1:0", comm.ServerConfig{})
		assert.NoError(t, err)
		lf, _, err := createLedgerFactory(conf, &disabled.Provider{})
		assert.NoError(t, err)
		registrar := initializeMultichannelRegistrar(
			nil,
			nil,
			&cluster.PredicateDialer{},
			comm.ServerConfig{},
			srv,
			conf,
			signer,
			&disabled.Provider{},
			&server_mocks.HealthChecker{},
			lf,
			cryptoProvider,
		)
		assert.NotNil(t, registrar)
		assert.Empty(t, registrar.SystemChannelID())
	})
}

func TestInitializeGrpcServer(t *testing.T) {
	// get a free random port
	listenAddr := func() string {
		l, _ := net.Listen("tcp", "localhost:0")
		l.Close()
		return l.Addr().String()
	}()
	host := strings.Split(listenAddr, ":")[0]
	port, _ := strconv.ParseUint(strings.Split(listenAddr, ":")[1], 10, 16)
	conf := &localconfig.TopLevel{
		General: localconfig.General{
			ListenAddress: host,
			ListenPort:    uint16(port),
			TLS: localconfig.TLS{
				Enabled:            false,
				ClientAuthRequired: false,
			},
		},
	}
	assert.NotPanics(t, func() {
		grpcServer := initializeGrpcServer(conf, initializeServerConfig(conf, nil))
		grpcServer.Listener().Close()
	})
}

// generateCryptoMaterials uses cryptogen to generate the necessary
// MSP files and TLS certificates
func generateCryptoMaterials(t *testing.T, cryptogen string) string {
	gt := NewGomegaWithT(t)
	cryptoPath := filepath.Join(tempDir, "crypto")

	cmd := exec.Command(
		cryptogen,
		"generate",
		"--config", filepath.Join(tempDir, "examplecom-config.yaml"),
		"--output", cryptoPath,
	)
	cryptogenProcess, err := gexec.Start(cmd, nil, nil)
	gt.Expect(err).NotTo(HaveOccurred())
	gt.Eventually(cryptogenProcess, time.Minute).Should(gexec.Exit(0))

	return cryptoPath
}

func TestUpdateTrustedRoots(t *testing.T) {
	cleanup := configtest.SetDevFabricConfigPath(t)
	defer cleanup()

	genesisFile := produceGenesisFile(t, genesisconfig.SampleDevModeSoloProfile, "testchannelid")
	defer os.Remove(genesisFile)

	cryptoPath := generateCryptoMaterials(t, cryptogen)
	defer os.RemoveAll(cryptoPath)

	// get a free random port
	listenAddr := func() string {
		l, _ := net.Listen("tcp", "localhost:0")
		l.Close()
		return l.Addr().String()
	}()
	port, _ := strconv.ParseUint(strings.Split(listenAddr, ":")[1], 10, 16)
	conf := &localconfig.TopLevel{
		General: localconfig.General{
			BootstrapMethod: "file",
			BootstrapFile:   genesisFile,
			ListenAddress:   "localhost",
			ListenPort:      uint16(port),
			TLS: localconfig.TLS{
				Enabled:            false,
				ClientAuthRequired: false,
			},
		},
	}
	grpcServer := initializeGrpcServer(conf, initializeServerConfig(conf, nil))
	caMgr := &caManager{
		appRootCAsByChain:     make(map[string][][]byte),
		ordererRootCAsByChain: make(map[string][][]byte),
	}
	callback := func(bundle *channelconfig.Bundle) {
		if grpcServer.MutualTLSRequired() {
			t.Log("callback called")
			caMgr.updateTrustedRoots(bundle, grpcServer)
		}
	}
	lf, _, err := createLedgerFactory(conf, &disabled.Provider{})
	assert.NoError(t, err)
	bootBlock := file.New(genesisFile).GenesisBlock()
	initializeBootstrapChannel(bootBlock, lf)
	signer := &server_mocks.SignerSerializer{}

	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)

	initializeMultichannelRegistrar(
		bootBlock,
		onboarding.NewReplicationInitiator(lf, bootBlock, conf, comm.SecureOptions{}, signer, cryptoProvider),
		&cluster.PredicateDialer{},
		comm.ServerConfig{},
		nil,
		genesisConfig(t, genesisFile),
		signer,
		&disabled.Provider{},
		&server_mocks.HealthChecker{},
		lf,
		cryptoProvider,
		callback,
	)
	t.Logf("# app CAs: %d", len(caMgr.appRootCAsByChain["testchannelid"]))
	t.Logf("# orderer CAs: %d", len(caMgr.ordererRootCAsByChain["testchannelid"]))
	// mutual TLS not required so no updates should have occurred
	assert.Equal(t, 0, len(caMgr.appRootCAsByChain["testchannelid"]))
	assert.Equal(t, 0, len(caMgr.ordererRootCAsByChain["testchannelid"]))
	grpcServer.Listener().Close()

	conf = &localconfig.TopLevel{
		General: localconfig.General{
			ListenAddress: "localhost",
			ListenPort:    uint16(port),
			TLS: localconfig.TLS{
				Enabled:            true,
				ClientAuthRequired: true,
				PrivateKey:         filepath.Join(cryptoPath, "ordererOrganizations", "example.com", "orderers", "127.0.0.1.example.com", "tls", "server.key"),
				Certificate:        filepath.Join(cryptoPath, "ordererOrganizations", "example.com", "orderers", "127.0.0.1.example.com", "tls", "server.crt"),
			},
		},
	}
	grpcServer = initializeGrpcServer(conf, initializeServerConfig(conf, nil))
	caMgr = &caManager{
		appRootCAsByChain:     make(map[string][][]byte),
		ordererRootCAsByChain: make(map[string][][]byte),
	}

	clusterConf := initializeClusterClientConfig(conf)
	predDialer := &cluster.PredicateDialer{
		Config: clusterConf,
	}

	callback = func(bundle *channelconfig.Bundle) {
		if grpcServer.MutualTLSRequired() {
			t.Log("callback called")
			caMgr.updateTrustedRoots(bundle, grpcServer)
			caMgr.updateClusterDialer(predDialer, clusterConf.SecOpts.ServerRootCAs)
		}
	}
	initializeMultichannelRegistrar(
		bootBlock,
		onboarding.NewReplicationInitiator(lf, bootBlock, conf, comm.SecureOptions{}, signer, cryptoProvider),
		predDialer,
		comm.ServerConfig{},
		nil,
		genesisConfig(t, genesisFile),
		signer,
		&disabled.Provider{},
		&server_mocks.HealthChecker{},
		lf,
		cryptoProvider,
		callback,
	)
	t.Logf("# app CAs: %d", len(caMgr.appRootCAsByChain["testchannelid"]))
	t.Logf("# orderer CAs: %d", len(caMgr.ordererRootCAsByChain["testchannelid"]))
	// mutual TLS is required so updates should have occurred
	// we expect an intermediate and root CA for apps and orderers
	assert.Equal(t, 2, len(caMgr.appRootCAsByChain["testchannelid"]))
	assert.Equal(t, 2, len(caMgr.ordererRootCAsByChain["testchannelid"]))
	assert.Len(t, predDialer.Config.SecOpts.ServerRootCAs, 2)
	grpcServer.Listener().Close()
}

func TestRootServerCertAggregation(t *testing.T) {
	caMgr := &caManager{
		appRootCAsByChain:     make(map[string][][]byte),
		ordererRootCAsByChain: make(map[string][][]byte),
	}

	predDialer := &cluster.PredicateDialer{
		Config: comm.ClientConfig{},
	}

	ca1, err := tlsgen.NewCA()
	require.NoError(t, err)

	ca2, err := tlsgen.NewCA()
	require.NoError(t, err)

	caMgr.ordererRootCAsByChain["foo"] = [][]byte{ca1.CertBytes()}
	caMgr.ordererRootCAsByChain["bar"] = [][]byte{ca1.CertBytes()}

	caMgr.updateClusterDialer(predDialer, [][]byte{ca2.CertBytes(), ca2.CertBytes(), ca2.CertBytes()})

	require.Len(t, predDialer.Config.SecOpts.ServerRootCAs, 2)
	require.Contains(t, predDialer.Config.SecOpts.ServerRootCAs, ca1.CertBytes())
	require.Contains(t, predDialer.Config.SecOpts.ServerRootCAs, ca2.CertBytes())
}

func TestConfigureClusterListener(t *testing.T) {
	logEntries := make(chan string, 100)

	allocatePort := func() uint16 {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		assert.NoError(t, err)
		_, portStr, err := net.SplitHostPort(l.Addr().String())
		assert.NoError(t, err)
		port, err := strconv.ParseInt(portStr, 10, 64)
		assert.NoError(t, err)
		assert.NoError(t, l.Close())
		t.Log("picked unused port", port)
		return uint16(port)
	}

	unUsedPort := allocatePort()

	backupLogger := logger
	logger = logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		logEntries <- entry.Message
		return nil
	}))

	defer func() {
		logger = backupLogger
	}()

	ca, err := tlsgen.NewCA()
	assert.NoError(t, err)
	serverKeyPair, err := ca.NewServerCertKeyPair("127.0.0.1")
	assert.NoError(t, err)

	loadPEM := func(fileName string) ([]byte, error) {
		switch fileName {
		case "cert":
			return serverKeyPair.Cert, nil
		case "key":
			return serverKeyPair.Key, nil
		case "ca":
			return ca.CertBytes(), nil
		default:
			return nil, errors.New("I/O error")
		}
	}

	for _, testCase := range []struct {
		name               string
		conf               *localconfig.TopLevel
		generalConf        comm.ServerConfig
		generalSrv         *comm.GRPCServer
		shouldBeEqual      bool
		expectedPanic      string
		expectedLogEntries []string
	}{
		{
			name:        "invalid certificate",
			generalConf: comm.ServerConfig{},
			conf: &localconfig.TopLevel{
				General: localconfig.General{
					Cluster: localconfig.Cluster{
						ListenAddress:     "127.0.0.1",
						ListenPort:        5000,
						ServerPrivateKey:  "key",
						ServerCertificate: "bad",
						RootCAs:           []string{"ca"},
					},
				},
			},
			expectedPanic:      "Failed to load cluster server certificate from 'bad' (I/O error)",
			generalSrv:         &comm.GRPCServer{},
			expectedLogEntries: []string{"Failed to load cluster server certificate from 'bad' (I/O error)"},
		},
		{
			name:        "invalid key",
			generalConf: comm.ServerConfig{},
			conf: &localconfig.TopLevel{
				General: localconfig.General{
					Cluster: localconfig.Cluster{
						ListenAddress:     "127.0.0.1",
						ListenPort:        5000,
						ServerPrivateKey:  "bad",
						ServerCertificate: "cert",
						RootCAs:           []string{"ca"},
					},
				},
			},
			expectedPanic:      "Failed to load cluster server key from 'bad' (I/O error)",
			generalSrv:         &comm.GRPCServer{},
			expectedLogEntries: []string{"Failed to load cluster server key from 'bad' (I/O error)"},
		},
		{
			name:        "invalid ca cert",
			generalConf: comm.ServerConfig{},
			conf: &localconfig.TopLevel{
				General: localconfig.General{
					Cluster: localconfig.Cluster{
						ListenAddress:     "127.0.0.1",
						ListenPort:        5000,
						ServerPrivateKey:  "key",
						ServerCertificate: "cert",
						RootCAs:           []string{"bad"},
					},
				},
			},
			expectedPanic:      "Failed to load CA cert file 'bad' (I/O error)",
			generalSrv:         &comm.GRPCServer{},
			expectedLogEntries: []string{"Failed to load CA cert file 'bad' (I/O error)"},
		},
		{
			name:        "bad listen address",
			generalConf: comm.ServerConfig{},
			conf: &localconfig.TopLevel{
				General: localconfig.General{
					Cluster: localconfig.Cluster{
						ListenAddress:     "99.99.99.99",
						ListenPort:        unUsedPort,
						ServerPrivateKey:  "key",
						ServerCertificate: "cert",
						RootCAs:           []string{"ca"},
					},
				},
			},
			expectedPanic: fmt.Sprintf("Failed creating gRPC server on 99.99.99.99:%d due "+
				"to listen tcp 99.99.99.99:%d:", unUsedPort, unUsedPort),
			generalSrv: &comm.GRPCServer{},
		},
		{
			name:        "green path",
			generalConf: comm.ServerConfig{},
			conf: &localconfig.TopLevel{
				General: localconfig.General{
					Cluster: localconfig.Cluster{
						ListenAddress:     "127.0.0.1",
						ListenPort:        5000,
						ServerPrivateKey:  "key",
						ServerCertificate: "cert",
						RootCAs:           []string{"ca"},
					},
				},
			},
			generalSrv: &comm.GRPCServer{},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.shouldBeEqual {
				conf, srv := configureClusterListener(testCase.conf, testCase.generalConf, loadPEM)
				assert.Equal(t, conf, testCase.generalConf)
				assert.Equal(t, srv, testCase.generalSrv)
			}

			if testCase.expectedPanic != "" {
				f := func() {
					configureClusterListener(testCase.conf, testCase.generalConf, loadPEM)
				}
				assert.Contains(t, panicMsg(f), testCase.expectedPanic)
			} else {
				configureClusterListener(testCase.conf, testCase.generalConf, loadPEM)
			}
			// Ensure logged messages that are expected were all logged
			var loggedMessages []string
			for len(logEntries) > 0 {
				logEntry := <-logEntries
				loggedMessages = append(loggedMessages, logEntry)
			}
			assert.Subset(t, loggedMessages, testCase.expectedLogEntries)
		})
	}
}

func TestReuseListener(t *testing.T) {
	t.Run("good to reuse", func(t *testing.T) {
		top := &localconfig.TopLevel{General: localconfig.General{TLS: localconfig.TLS{Enabled: true}}}
		require.True(t, reuseListener(top))
	})

	t.Run("reuse tls disabled", func(t *testing.T) {
		top := &localconfig.TopLevel{}
		require.PanicsWithValue(
			t,
			"TLS is required for running ordering nodes of cluster type.",
			func() { reuseListener(top) },
		)
	})

	t.Run("good not to reuse", func(t *testing.T) {
		top := &localconfig.TopLevel{
			General: localconfig.General{
				Cluster: localconfig.Cluster{
					ListenAddress:     "127.0.0.1",
					ListenPort:        5000,
					ServerPrivateKey:  "key",
					ServerCertificate: "bad",
				},
			},
		}
		require.False(t, reuseListener(top))
	})

	t.Run("partial config", func(t *testing.T) {
		top := &localconfig.TopLevel{
			General: localconfig.General{
				Cluster: localconfig.Cluster{
					ListenAddress:     "127.0.0.1",
					ListenPort:        5000,
					ServerCertificate: "bad",
				},
			},
		}
		require.PanicsWithValue(
			t,
			"Options: General.Cluster.ListenPort, General.Cluster.ListenAddress,"+
				" General.Cluster.ServerCertificate, General.Cluster.ServerPrivateKey, should be defined altogether.",
			func() { reuseListener(top) },
		)
	})
}

func TestInitializeEtcdraftConsenter(t *testing.T) {
	consenters := make(map[string]consensus.Consenter)

	tmpdir, err := ioutil.TempDir("", "main_test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	rlf, err := fileledger.New(tmpdir, &disabled.Provider{})
	require.NoError(t, err)

	conf := genesisconfig.Load(genesisconfig.SampleInsecureSoloProfile, configtest.GetDevConfigDir())
	genesisBlock := encoder.New(conf).GenesisBlock()

	ca, _ := tlsgen.NewCA()
	crt, _ := ca.NewServerCertKeyPair("127.0.0.1")

	srv, err := comm.NewGRPCServer("127.0.0.1:0", comm.ServerConfig{})
	assert.NoError(t, err)

	cryptoProvider, err := sw.NewDefaultSecurityLevelWithKeystore(sw.NewDummyKeyStore())
	assert.NoError(t, err)

	initializeEtcdraftConsenter(
		consenters,
		&localconfig.TopLevel{},
		rlf,
		&cluster.PredicateDialer{},
		genesisBlock,
		onboarding.NewReplicationInitiator(rlf, genesisBlock, nil, comm.SecureOptions{}, nil, cryptoProvider),
		comm.ServerConfig{
			SecOpts: comm.SecureOptions{
				Certificate: crt.Cert,
				Key:         crt.Key,
				UseTLS:      true,
			},
		},
		srv,
		&multichannel.Registrar{},
		&disabled.Provider{},
		cryptoProvider,
	)
	assert.NotNil(t, consenters["etcdraft"])
}

func genesisConfig(t *testing.T, genesisFile string) *localconfig.TopLevel {
	t.Helper()
	localMSPDir := configtest.GetDevMspDir()
	return &localconfig.TopLevel{
		General: localconfig.General{
			BootstrapMethod: "file",
			BootstrapFile:   genesisFile,
			LocalMSPDir:     localMSPDir,
			LocalMSPID:      "SampleOrg",
			BCCSP: &factory.FactoryOpts{
				ProviderName: "SW",
				SwOpts: &factory.SwOpts{
					HashFamily: "SHA2",
					SecLevel:   256,
				},
			},
		},
	}
}

func panicMsg(f func()) string {
	var message interface{}
	func() {

		defer func() {
			message = recover()
		}()

		f()

	}()

	return message.(string)

}

func produceGenesisFile(t *testing.T, profile, channelID string) string {
	conf := genesisconfig.Load(profile, configtest.GetDevConfigDir())
	f, err := ioutil.TempFile("", fmt.Sprintf("%s-genesis_block-", t.Name()))
	require.NoError(t, err)
	_, err = f.Write(protoutil.MarshalOrPanic(encoder.New(conf).GenesisBlockForChannel(channelID)))
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)
	return f.Name()
}
