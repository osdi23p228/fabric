/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"io/ioutil"

	"github.com/osdi23p228/fabric/common/ledger/blockledger"
	"github.com/osdi23p228/fabric/common/ledger/blockledger/fileledger"
	"github.com/osdi23p228/fabric/common/metrics"
	config "github.com/osdi23p228/fabric/orderer/common/localconfig"
	"github.com/pkg/errors"
)

func createLedgerFactory(conf *config.TopLevel, metricsProvider metrics.Provider) (blockledger.Factory, string, error) {
	ld := conf.FileLedger.Location
	var err error
	if ld == "" {
		if ld, err = ioutil.TempDir("", conf.FileLedger.Prefix); err != nil {
			logger.Panic("Error creating temp dir:", err)
		}
	}

	logger.Debug("Ledger dir:", ld)
	lf, err := fileledger.New(ld, metricsProvider)
	if err != nil {
		return nil, "", errors.WithMessage(err, "Error in opening ledger factory")
	}
	return lf, ld, nil
}
