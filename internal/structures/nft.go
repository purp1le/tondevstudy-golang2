package structures

import (
	"github.com/xssnick/tonutils-go/address"
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


type NftSaleData struct {
	
}