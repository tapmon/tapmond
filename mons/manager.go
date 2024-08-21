package mons

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/lightninglabs/lndclient"
	"github.com/lightninglabs/taproot-assets/taprpc"
	"github.com/lightninglabs/taproot-assets/taprpc/mintrpc"
	"github.com/lightningnetwork/lnd/chainntnfs"
	"google.golang.org/grpc"
)

var (
	tapmonV1NameFormatting = "tapmon-V1-%v"
)

type MonMetadata struct {
	MonVersion uint32 `json:"mon_version"`
}

type Manager struct {
	tapClient     taprpc.TaprootAssetsClient
	mintClient    mintrpc.MintClient
	chainNotifier lndclient.ChainNotifierClient
	chainKit      lndclient.ChainKitClient
}

func NewManager(tapdConn *grpc.ClientConn, chainNotifier lndclient.ChainNotifierClient,
	chainKit lndclient.ChainKitClient) *Manager {

	log.Debug("Creating new mon manager")

	return &Manager{
		tapClient:     taprpc.NewTaprootAssetsClient(tapdConn),
		mintClient:    mintrpc.NewMintClient(tapdConn),
		chainNotifier: chainNotifier,
		chainKit:      chainKit,
	}
}

// IndexMons will index all monsters
func (m *Manager) IndexMons() error {

	return nil
}

// MintMon will mint a new monster
func (m *Manager) MintMon(ctx context.Context, name string) (*Mon, error) {
	monMetadata := &MonMetadata{
		MonVersion: 1,
	}
	// Create the json string of the metadata
	monMetadataBytes, err := json.Marshal(monMetadata)
	if err != nil {
		return nil, err
	}
	_, err = m.mintClient.MintAsset(
		ctx, &mintrpc.MintAssetRequest{
			Asset: &mintrpc.MintAsset{
				Name:         fmt.Sprintf(tapmonV1NameFormatting, name),
				AssetVersion: taprpc.AssetVersion_ASSET_VERSION_V1,
				AssetMeta: &taprpc.AssetMeta{
					Type: taprpc.AssetMetaType_META_TYPE_JSON,
					Data: monMetadataBytes,
				},
				Amount:    1,
				AssetType: taprpc.AssetType_COLLECTIBLE,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	finalizeRes, err := m.mintClient.FinalizeBatch(
		ctx, &mintrpc.FinalizeBatchRequest{},
	)
	if err != nil {
		return nil, err
	}
	log.Debugf("Batch txid: %v\n", finalizeRes.Batch.BatchTxid)

	batchHash, err := chainhash.NewHashFromStr(finalizeRes.Batch.BatchTxid)
	if err != nil {
		return nil, err
	}
	// We'll wait for the transaction to have 2 confirmations before
	// returning.
	confChan, errChan, err := m.chainNotifier.RegisterConfirmationsNtfn(
		ctx, batchHash, nil, 2, int32(finalizeRes.Batch.HeightHint),
	)
	if err != nil {
		return nil, err
	}
	var conf *chainntnfs.TxConfirmation
	select {
	case <-errChan:
		return nil, err
	case conf = <-confChan:
		log.Debugf("Batch txid %v has been confirmed in block %v\n",
			finalizeRes.Batch.BatchTxid, conf.Block)
	}

	// Get the resulting Mon
	mintedMon, err := GenerateMonster(conf.BlockHash, batchHash)
	if err != nil {
		return nil, err
	}
	log.Debugf("Minted Mon: %v\n", mintedMon)

	return mintedMon, nil
}
