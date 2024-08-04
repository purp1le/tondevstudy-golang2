package main

import (
	"context"
	"encoding/hex"
	"ton-lessons2/internal/app"
	"ton-lessons2/internal/structures"

	"github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
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
		return err
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


	if err := DeployNewNft(
		wall,
		api,
		"EQBZ6u2yK7um2oEqBU1M0t92N_jjRctMd7SfvNixPYtYJOEj",
	); err != nil {
		return err
	}

	return nil
}

// collection content storage
// https://example.com/nft/collection.json
// https://example.com/nft/item/1.json
// https://example.com/nft/item/2.json
// https://example.com/nft/item/3.json
// https://example.com/nft/item/ -- collection
// 3.json, 2.json -- nft item

func DeployNewNft(wall *wallet.Wallet, api *ton.APIClient, collectionAddr string) error {
	collectionAddress := address.MustParseAddr(collectionAddr)

	collection := nft.NewCollectionClient(api, collectionAddress)

	collectionData, err := collection.GetCollectionData(context.Background())
	if err != nil {
		return err
	}

	nftAddr, err := collection.GetNFTAddressByIndex(
		context.Background(),
		collectionData.NextItemIndex,
	)
	if err != nil {
		return err
	}

	mintData, err := collection.BuildMintEditablePayload(
		collectionData.NextItemIndex,
		wall.Address(),
		wall.Address(),
		tlb.MustFromTON("0.01"),
		&nft.ContentOffchain{
			"https://s.getgems.io/nft/b/c/6627eb103ccea4453c4073d1/rev/4666.dcd2sb.json",
		},
	)
	if err != nil {
		return err
	}

	transaction, block, err := wall.SendWaitTransaction(
		context.Background(),
		wallet.SimpleMessage(
			address.MustParseAddr(collectionAddr),
			tlb.MustFromTON("0.025"),
			mintData,
		),
	)

	logrus.Info("NFT ADDRESS - ", nftAddr)
	logrus.Info("TX HASH - ", hex.EncodeToString(transaction.Hash))
	logrus.Info("BLOCK SEQNO - ", block.SeqNo)


	return nil
}

func ChangeNFTMetadata(wall *wallet.Wallet, collectionAddr string, uri string, commontContent string) error {
	collectionContent := nft.ContentOffchain{
		URI: uri,
	}

	collectionContentCell, err := collectionContent.ContentCell()
	if err != nil {
		return err
	}

	commonContentCell := cell.BeginCell().MustStoreStringSnake(commontContent).EndCell()

	content := cell.BeginCell().
		MustStoreRef(collectionContentCell).
		MustStoreRef(commonContentCell).EndCell()

	royalty := structures.RoyaltyParams{
		Factor:  10,
		Base:    100,
		Address: wall.Address(),
	}

	royaltyCell, err := tlb.ToCell(&royalty)
	if err != nil {
		return err
	}

	changeBody := cell.BeginCell().
		MustStoreUInt(4, 32).
		MustStoreUInt(0, 64).
		MustStoreRef(content).
		MustStoreRef(royaltyCell).EndCell()

	transaction, block, err := wall.SendWaitTransaction(
		context.Background(),
		wallet.SimpleMessage(
			address.MustParseAddr(collectionAddr),
			tlb.MustFromTON("0.05"),
			changeBody,
		),
	)

	logrus.Info("TX HASH - ", hex.EncodeToString(transaction.Hash))
	logrus.Info("BLOCK SEQNO - ", block.SeqNo)
	return nil
}

func DeployNewCollection(wall *wallet.Wallet) error {
	collectionCode, err := GetNftCollectionCode()
	if err != nil {
		return err
	}

	collectionData, err := GetNftCollectionData(
		wall.Address(),
		cell.BeginCell().EndCell(),
		10,
		100,
		wall.Address(),
	)
	if err != nil {
		return err
	}

	contractAddr, transaction, block, err := wall.DeployContractWaitTransaction(
		context.Background(),
		tlb.MustFromTON("0.2"),
		nil,
		collectionCode,
		collectionData,
	)
	if err != nil {
		return err
	}

	logrus.Info("CONTRACT ADDRESS - ", contractAddr)
	logrus.Info("TX HASH - ", hex.EncodeToString(transaction.Hash))
	logrus.Info("BLOCK SEQNO - ", block.SeqNo)

	return nil
}

func GetNftCollectionCode() (*cell.Cell, error) {
	hexByteCode := "b5ee9c720102140100021f000114ff00f4a413f4bcf2c80b0102016202030202cd04050201200e0f04e7d10638048adf000e8698180b8d848adf07d201800e98fe99ff6a2687d20699fea6a6a184108349e9ca829405d47141baf8280e8410854658056b84008646582a802e78b127d010a65b509e58fe59f80e78b64c0207d80701b28b9e382f970c892e000f18112e001718112e001f181181981e0024060708090201200a0b00603502d33f5313bbf2e1925313ba01fa00d43028103459f0068e1201a44343c85005cf1613cb3fccccccc9ed54925f05e200a6357003d4308e378040f4966fa5208e2906a4208100fabe93f2c18fde81019321a05325bbf2f402fa00d43022544b30f00623ba9302a402de04926c21e2b3e6303250444313c85005cf1613cb3fccccccc9ed54002c323401fa40304144c85005cf1613cb3fccccccc9ed54003c8e15d4d43010344130c85005cf1613cb3fccccccc9ed54e05f04840ff2f00201200c0d003d45af0047021f005778018c8cb0558cf165004fa0213cb6b12ccccc971fb008002d007232cffe0a33c5b25c083232c044fd003d0032c03260001b3e401d3232c084b281f2fff2742002012010110025bc82df6a2687d20699fea6a6a182de86a182c40043b8b5d31ed44d0fa40d33fd4d4d43010245f04d0d431d430d071c8cb0701cf16ccc980201201213002fb5dafda89a1f481a67fa9a9a860d883a1a61fa61ff480610002db4f47da89a1f481a67fa9a9a86028be09e008e003e00b0"
	byteCode, err := hex.DecodeString(hexByteCode)
	if err != nil {
		return nil, err
	}

	codeCell, err := cell.FromBOC(byteCode)
	if err != nil {
		return nil, err
	}

	return codeCell, nil
}

func GetNftItemCode() (*cell.Cell, error) {
	hexByteCode := "b5ee9c72010212010002e5000114ff00f4a413f4bcf2c80b0102016202030202ce0405020120101102012006070201200e0f04f70c8871c02497c0f83434c0c05c6c2497c0f83e903e900c7e800c5c75c87e800c7e800c3c00816ce38596db088d148cb1c17cb865407e90353e900c040d3c00f801f4c7f4cfe08417f30f45148c2ea3a28c8412040dc409841140b820840bf2c9a8948c2eb8c0a0840701104a948c2ea3a28c8412040dc409841140a008090a0b00113e910c1c2ebcb8536001f65136c705f2e191fa4021f001fa40d20031fa00820afaf0801ca121945315a0a1de22d70b01c300209206a19136e220c2fff2e192218e3e821005138d91c8500acf16500ccf1671244a145446b0708010c8cb055007cf165005fa0215cb6a12cb1fcb3f226eb39458cf17019132e201c901fb00105894102b385be20c0080135f03333334347082108b77173504c8cbff58cf164430128040708010c8cb055007cf165005fa0215cb6a12cb1fcb3f226eb39458cf17019132e201c901fb0001f65134c705f2e191fa4021f001fa40d20031fa00820afaf0801ca121945315a0a1de22d70b01c300209206a19136e220c2fff2e192218e3e8210511a4463c85008cf16500ccf1671244814544690708010c8cb055007cf165005fa0215cb6a12cb1fcb3f226eb39458cf17019132e201c901fb00103894102b365be20d0046e03136373782101a0b9d5116ba9e5131c705f2e19a01d4304400f003e05f06840ff2f00082028e3527f0018210d53276db103845006d71708010c8cb055007cf165005fa0215cb6a12cb1fcb3f226eb39458cf17019132e201c901fb0093303335e25503f0030082028e3527f0018210d53276db103848006d71708010c8cb055007cf165005fa0215cb6a12cb1fcb3f226eb39458cf17019132e201c901fb0093303630e25503f00300413b513434cffe900835d27080271fc07e90353e900c040d440d380c1c165b5b5b600025013232cfd400f3c58073c5b30073c5b27b5520000dbf03a78013628c000bbc7e7f801184"
	byteCode, err := hex.DecodeString(hexByteCode)
	if err != nil {
		return nil, err
	}

	codeCell, err := cell.FromBOC(byteCode)
	if err != nil {
		return nil, err
	}

	return codeCell, nil
}

func GetNftCollectionData(
	owner *address.Address,
	content *cell.Cell,
	factor uint16,
	base uint16,
	royalyAddress *address.Address,
) (*cell.Cell, error) {
	nftItemCode, err := GetNftItemCode()
	if err != nil {
		return nil, err
	}
	collectionData := structures.NftCollectionData{
		Owner:         owner,
		NextTimeIndex: 0,
		Content:       content,
		NftItemCode:   nftItemCode,
		RoyaltyParams: structures.RoyaltyParams{
			Factor:  factor,
			Base:    base,
			Address: royalyAddress,
		},
	}

	data, err := tlb.ToCell(&collectionData)
	if err != nil {
		return nil, err
	}

	return data, nil
}
