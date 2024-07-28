package structures

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TransferNotification struct {
	_          tlb.Magic        `tlb:"#7362d09c"`
	QueryId    uint64           `tlb:"## 64"`
	Amount     tlb.Coins        `tlb:"."`
	Sender     *address.Address `tlb:"addr"`
	FwdPayload *cell.Cell       `tlb:"either . ^"`
}

type Comment struct {
	
}