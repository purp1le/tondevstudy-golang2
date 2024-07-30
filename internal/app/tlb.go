package app

import (
	"ton-lessons2/internal/structures"

	"github.com/xssnick/tonutils-go/tlb"
)

func InitTlb() {
	tlb.Register(structures.DedustAssetJetton{})
	tlb.Register(structures.DedustAssetNative{})
}