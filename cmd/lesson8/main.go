package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"time"
	"ton-lessons2/internal/app"
	"ton-lessons2/internal/structures"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := app.InitApp(); err != nil {
		return nil
	}

	client := liteclient.NewConnectionPool()

	if err := client.AddConnectionsFromConfig(
		context.Background(),
		app.CFG.MainnetConfig,
	); err != nil {
		return err
	}

	api := ton.NewAPIClient(client)

	wall, err := wallet.FromSeed(api, app.CFG.Wallet.Seed, wallet.V4R2)
	if err != nil {
		return err
	}

	if err := SaleNftViaDeployer(
		wall,
		"EQDyhWPUky2fansDmTJl7m74JF0OTyXwoPUvFsGCHdXD_MF-",
		"UQDK5FxcwdHHRCMF9JFuPnD1z5_Ioxv8BjjVQhWSw3k__MXO",
	); err != nil {
		return err
	}

	return nil
}

func SaleNftViaDeployer(
	wall *wallet.Wallet,
	deployerAddress string,
	nftAddress string,
) error {
	code, err := getSaleCode()
	if err != nil {
		return err
	}

	data, err := getSaleData(
		wall.Address().String(),
		nftAddress,
		wall.Address().String(),
		decimal.NewFromFloat32(1e8),
	)
	if err != nil {
		return err
	}

	stateInit := tlb.StateInit{
		Code: code,
		Data: data,
	}

	stateInitCell, err := tlb.ToCell(&stateInit)
	if err != nil {
		return err
	}

	transferRequestCell := cell.BeginCell().
		MustStoreUInt(0x5fcc3d14, 32).
		MustStoreUInt(uint64(rand.Uint32()), 64).
		MustStoreAddr(address.MustParseAddr(deployerAddress)).
		MustStoreAddr(wall.Address()).
		MustStoreUInt(0, 1).
		MustStoreBigCoins(tlb.MustFromTON("0.05").Nano()).
		MustStoreUInt(0x0fe0ede, 32).
		MustStoreRef(stateInitCell).
		MustStoreRef(cell.BeginCell().EndCell()).EndCell()


	transaction, block, err := wall.SendWaitTransaction(
		context.Background(),
		wallet.SimpleMessage(address.MustParseAddr(nftAddress), tlb.MustFromTON("0.15"), transferRequestCell),
	)
	if err != nil {
		return err
	}
	logrus.Info("TX HASH - ", hex.EncodeToString(transaction.Hash))
	logrus.Info("BLOCK SEQNO - ", block.SeqNo)

	return nil
}

func TransferNft(
	wall *wallet.Wallet,
	nftAddr string,
	to string,
) error {
	transferRequest := structures.NftTransferRequest{
		QueryId:             uint64(rand.Uint32()),
		NewOwner:            address.MustParseAddr(to),
		ResponseDestination: wall.Address(),
		CustomPayload:       nil,
		FwdAmount:           tlb.MustFromTON("0.05"),
		FwdPayload:          nil,
	}

	transferRequestCell, err := tlb.ToCell(&transferRequest)
	if err != nil {
		return nil
	}

	transaction, block, err := wall.SendWaitTransaction(
		context.Background(),
		wallet.SimpleMessage(address.MustParseAddr(nftAddr), tlb.MustFromTON("0.15"), transferRequestCell),
	)
	if err != nil {
		return err
	}
	logrus.Info("TX HASH - ", hex.EncodeToString(transaction.Hash))
	logrus.Info("BLOCK SEQNO - ", block.SeqNo)

	return nil
}

func DeployNewSale(
	wall *wallet.Wallet,
) error {
	code, err := getSaleCode()
	if err != nil {
		return err
	}

	data, err := getSaleData(
		wall.Address().String(),
		wall.Address().String(),
		"EQC6Qowm60vi_mkuKtuimqqIFz2HxHUx0PSuE-ZoFkFTpnzi",
		decimal.NewFromFloat(1e8),
	)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	contractAddr, transaction, block, err := wall.DeployContractWaitTransaction(
		context.Background(),
		tlb.MustFromTON("0.1"),
		nil,
		code,
		data,
	)
	if err != nil {
		return err
	}

	logrus.Info("CONTRACT ADDRESS - ", contractAddr)
	logrus.Info("TX HASH - ", hex.EncodeToString(transaction.Hash))
	logrus.Info("BLOCK SEQNO - ", block.SeqNo)
	return nil
}

func getSaleCode() (*cell.Cell, error) {
	saleCode := "b5ee9c7201020f01000393000114ff00f4a413f4bcf2c80b0102016202030202cd04050201200d0e02f7d00e8698180b8d8492f82707d201876a2686980698ffd207d207d207d006a698fe99f9818382985638060004a9885698f85ef10e1804a1805699fc708c5b31b0b731b2b64166382c939996f2805f115e000c92f877012eba4e10116408115dd15e0009159d8d829e4e382d87181156000f968ca164108363610405d4060701d166084017d7840149828148c2fbcb87089343e903e803e903e800c14e4a848685421e845a814a4087e9116dc20043232c15400f3c5807e80b2dab25c7ec00970800975d27080ac2386d411487e9116dc20043232c15400f3c5807e80b2dab25c7ec00408e48d0d3896a0c006430316cb2d430d0d307218020b0f2d19522c3008e14810258f8235341a1bc04f82302a0b913b0f2d1969132e201d43001fb0004f053c7c705b08e5d135f03323737373704fa00fa00fa00305321a121a1c101f2d19805d0fa40fa00fa40fa003030c83202cf1658fa0201cf165004fa02c97020104810371045103408c8cb0017cb1f5005cf165003cf1601cf1601fa02cccb1fcb3fc9ed54e0b3e30230313728c003e30228c000e30208c00208090a0b0086353b3b5374c705925f0be05173c705f2e1f4821005138d9118baf2e1f5fa403010481037553208c8cb0017cb1f5005cf165003cf1601cf1601fa02cccb1fcb3fc9ed5400e23839821005f5e10018bef2e1c95346c7055152c70515b1f2e1ca702082105fcc3d14218010c8cb0528cf1621fa02cb6acb1f15cb3f27cf1627cf1614ca0023fa0213ca00c98306fb0071705417005e331034102308c8cb0017cb1f5005cf165003cf1601cf1601fa02cccb1fcb3fc9ed54001836371038476514433070f005002098554410241023f005e05f0a840ff2f000ec21fa445b708010c8cb055003cf1601fa02cb6ac971fb00702082105fcc3d14c8cb1f5230cb3f24cf165004cf1613ca008209c9c380fa0212ca00c9718018c8cb0527cf1670fa02cb6acc25fa445bc98306fb00715560f8230108c8cb0017cb1f5005cf165003cf1601cf1601fa02cccb1fcb3fc9ed540087bce1676a2686980698ffd207d207d207d006a698fe99f982de87d207d007d207d001829a15090d0e080f968cc93fd222d937d222d91fd222dc1082324ac28056000aac040081bee5ef6a2686980698ffd207d207d207d006a698fe99f9801687d207d007d207d001829b15090d0e080f968cd14fd222d947d222d91fd222d85e00085881aaa894"
	saleCodeBytes, err := hex.DecodeString(saleCode)
	if err != nil {
		return nil, err
	}

	code, err := cell.FromBOC(saleCodeBytes)
	if err != nil {
		return nil, err
	}

	return code, nil
}

func getSaleData(
	marketplaceAddress string,
	nftAddress string,
	owner string,
	fullPrice decimal.Decimal,
) (*cell.Cell, error) {
	saleData := structures.NftSaleData{
		IsComplete:         false,
		CreatedAt:          uint32(time.Now().Unix()),
		MarketplaceAddress: address.MustParseAddr(marketplaceAddress),
		NftAddress:         address.MustParseAddr(nftAddress),
		NftOwnerAddress:    address.MustParseAddr(owner),
		FullPrice:          tlb.MustFromNano(fullPrice.BigInt(), 0),
		Fees: structures.SaleFees{
			MarketplaceFeeAddress: address.MustParseAddr(marketplaceAddress),
			MarketplaceFee:        tlb.MustFromTON("0"),
			RoyaltyAddress:        address.MustParseAddr(marketplaceAddress),
			Royalty:               tlb.MustFromTON("0"),
		},
	}

	saleDataCell, err := tlb.ToCell(&saleData)
	if err != nil {
		return nil, err
	}

	return saleDataCell, nil
}

func DeployDeployer(
	wall *wallet.Wallet,
) error {
	code, err := getDeployerCode()
	if err != nil {
		return err
	}

	data, err := getDeployerData(
		wall.Address(),
	)
	if err != nil {
		return err
	}

	contractAddr, transaction, block, err := wall.DeployContractWaitTransaction(
		context.Background(),
		tlb.MustFromTON("0.1"),
		nil,
		code,
		data,
	)
	if err != nil {
		return err
	}

	logrus.Info("CONTRACT ADDRESS - ", contractAddr)
	logrus.Info("TX HASH - ", hex.EncodeToString(transaction.Hash))
	logrus.Info("BLOCK SEQNO - ", block.SeqNo)

	return nil
}

func getDeployerCode() (*cell.Cell, error) {
	deployerCode := "te6cckEBBQEA8AABFP8A9KQT9LzyyAsBAaDTIMcAkl8E4AHQ0wMBcbCSXwTg+kAwAdMfghAFE42RUiC64wIzMyLAAZJfA+ACgQIruo4W7UTQ+kAwWMcF8uGT1DDQ0wfUMAH7AOBbhA/y8AIC/DHTP/pA0x+CCP4O3hK98tGU1NTRIfkAcMjKB8v/ydB3dIAYyMsFywIizxaCCTEtAPoCy2sTzMzJcfsAcCB0ghBfzD0UIoAYyMsFUAnPFiP6AhjLahfLHxXLPxXLAgHPFgHPFsoAIfoCygDJWaEggggPQkC5lIED6KDjDXD7AgMEAAwwgggPQkAACIMG+wAl44cc"
	deployerCodeBytes, err := base64.StdEncoding.DecodeString(deployerCode)
	if err != nil {
		return nil, err
	}

	code, err := cell.FromBOC(deployerCodeBytes)
	if err != nil {
		return nil, err
	}

	return code, nil
}

func getDeployerData(owner *address.Address) (*cell.Cell, error) {
	return cell.BeginCell().MustStoreAddr(owner).EndCell(), nil
}
