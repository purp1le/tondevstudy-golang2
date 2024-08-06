package structures

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// ;; default#_ royalty_factor:uint16 royalty_base:uint16 royalty_address:MsgAddress = RoyaltyParams;
// ;; storage#_ owner_address:MsgAddress next_item_index:uint64
// ;;           ^[collection_content:^Cell common_content:^Cell]
// ;;           nft_item_code:^Cell
// ;;           royalty_params:^RoyaltyParams
// ;;           = Storage;

type RoyaltyParams struct {
	Factor  uint16           `tlb:"## 16"`
	Base    uint16           `tlb:"## 16"`
	Address *address.Address `tlb:"addr"`
}

type NftCollectionData struct {
	Owner         *address.Address `tlb:"addr"`
	NextTimeIndex uint64           `tlb:"## 64"`
	Content       *cell.Cell       `tlb:"^"`
	NftItemCode   *cell.Cell       `tlb:"^"`
	RoyaltyParams RoyaltyParams    `tlb:"^"`
}

type SaleFees struct {
	MarketplaceFeeAddress *address.Address `tlb:"addr"`
	MarketplaceFee        tlb.Coins        `tlb:"."`
	RoyaltyAddress        *address.Address `tlb:"addr"`
	Royalty               tlb.Coins        `tlb:"."`
}
type NftSaleData struct {
	IsComplete         bool             `tlb:"bool"`
	CreatedAt          uint32           `tlb:"## 32"`
	MarketplaceAddress *address.Address `tlb:"addr"`
	NftAddress         *address.Address `tlb:"addr"`
	NftOwnerAddress    *address.Address `tlb:"addr"`
	FullPrice          tlb.Coins        `tlb:"."`
	Fees               SaleFees         `tlb:"^"`
	SoldAt             uint32           `tlb:"## 32"`
	QueryId            uint64           `tlb:"## 64"`
}

// transfer#5fcc3d14 query_id:uint64 new_owner:MsgAddress response_destination:MsgAddress custom_payload:(Maybe ^Cell) forward_amount:(VarUInteger 16) forward_payload:(Either Cell ^Cell) = InternalMsgBody;
type NftTransferRequest struct {
	_                   tlb.Magic        `tlb:"#5fcc3d14"`
	QueryId             uint64           `tlb:"## 64"`
	NewOwner            *address.Address `tlb:"addr"`
	ResponseDestination *address.Address `tlb:"addr"`
	CustomPayload       *cell.Cell       `tlb:"maybe ^"`
	FwdAmount           tlb.Coins        `tlb:"."`
	FwdPayload          *cell.Cell       `tlb:"either . ^"`
}
