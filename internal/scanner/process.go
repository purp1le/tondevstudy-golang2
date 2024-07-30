package scanner

import (
	"fmt"
	"time"
	"ton-lessons2/internal/storage"
	"ton-lessons2/internal/structures"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"gorm.io/gorm"
)

func (s *scanner) processTransaction(
	trans *tlb.Transaction,
	dbtx *gorm.DB,
) error {
	if trans.IO.Out == nil {
		return nil
	}

	outMsgs, err := trans.IO.Out.ToSlice()
	if err != nil {
		return nil
	}

	for _, out := range outMsgs {
		if out.MsgType != tlb.MsgTypeExternalOut {
			continue
		}

		externalOut := out.AsExternalOut()
		if externalOut.Body == nil {
			continue
		}

		var dedustSwapEvent structures.DedustSwapEvent

		if err := tlb.LoadFromCell(
			&dedustSwapEvent,
			externalOut.Body.BeginParse(),
		); err != nil {
			continue
		}

		var (
			amountIn  string
			amountOut string
		)

		if dedustSwapEvent.AssetIn.Type() == "native" {
			amountIn = dedustSwapEvent.AmountIn.String() + " TON"
		} else {
			jettonAddr := dedustSwapEvent.AssetIn.AsJetton()
			amountIn = fmt.Sprintf("%s JETTON root [%s]",
				dedustSwapEvent.AmountIn.String(),
				address.NewAddress(0, byte(jettonAddr.WorkchainID), jettonAddr.AddressData).String(),
			)
		}

		if dedustSwapEvent.AssetOut.Type() == "native" {
			amountOut = dedustSwapEvent.AmountOut.String() + " TON"
		} else {
			jettonAddr := dedustSwapEvent.AssetOut.AsJetton()
			amountOut = fmt.Sprintf("%s JETTON root [%s]",
				dedustSwapEvent.AmountOut.String(),
				address.NewAddress(0, byte(jettonAddr.WorkchainID), jettonAddr.AddressData).String(),
			)
		}

		logrus.Info("[DDST] new swap!")
		logrus.Info("[DDST] swap from: ", dedustSwapEvent.ExtraInfo.SenderAddr)
		logrus.Info("[DDST] amount input: ", amountIn)
		logrus.Info("[DDST] amount input: ", amountOut)

		dedustSwap := storage.DedustSwap{
			PoolAddress:   externalOut.SrcAddr.String(),
			AssetIn:       dedustSwapEvent.AssetIn.Type(),
			AmountIn:      decimal.NewFromBigInt(dedustSwapEvent.AmountIn.Nano(), 0),
			AssetOut:      dedustSwapEvent.AssetOut.Type(),
			AmountOut:     decimal.NewFromBigInt(dedustSwapEvent.AmountOut.Nano(), 0),
			SenderAddress: dedustSwapEvent.ExtraInfo.SenderAddr.String(),
			Reserve0:      decimal.NewFromBigInt(dedustSwapEvent.ExtraInfo.Reserve0.Nano(), 0),
			Reserve1:      decimal.NewFromBigInt(dedustSwapEvent.ExtraInfo.Reserve1.Nano(), 0),
			CreatedAt:     time.Unix(int64(externalOut.CreatedAt), 0),
			ProcessedAt:   time.Now(),
		}

		if err := dbtx.Create(&dedustSwap).Error; err != nil {
			return err
		}
	}

	return nil
}
