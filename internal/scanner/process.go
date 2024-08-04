package scanner

import (
	"context"
	"encoding/hex"
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

func (s *scanner) processTransactionDedust(
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

func (s *scanner) processTransaction(
	trans *tlb.Transaction,
	dbtx *gorm.DB,
	master *tlb.BlockInfo,
) error {
	var (
		stonfiPart1 structures.StonfiSwapPart1
		stonfiPart2 structures.StonfiSwapPart2
		part2Found  = false
	)

	if trans.IO.In.MsgType != tlb.MsgTypeInternal {
		return nil
	}

	inMessage := trans.IO.In.AsInternal()
	if inMessage.Body == nil {
		return nil
	}

	if inMessage.SrcAddr.String() !=
		address.MustParseAddr("EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt").String() {
		return nil
	}

	// hex hash - 82566ad72b6568fe7276437d3b0c911aab65ed701c13601941b2917305e81c11
	accountInfo, err := s.api.GetAccount(
		context.Background(),
		master,
		address.NewAddress(0, 0, trans.AccountAddr),
	)
	if err != nil {
		return err
	}

	if hex.EncodeToString(accountInfo.Code.Hash()) != "82566ad72b6568fe7276437d3b0c911aab65ed701c13601941b2917305e81c11" {
		logrus.Warn("[SCN] hex hash dont equals: ", address.NewAddress(0, 0, trans.AccountAddr))
		return nil
	} 

	if err := tlb.LoadFromCell(&stonfiPart1, inMessage.Body.BeginParse()); err != nil {
		return nil
	}

	if trans.IO.Out == nil {
		return nil
	}

	outMessages, err := trans.IO.Out.ToSlice()
	if err != nil {
		return nil
	}

	for _, outMessage := range outMessages {
		if part2Found {
			continue
		}

		if outMessage.MsgType != tlb.MsgTypeInternal {
			continue
		}

		outInternalMessage := outMessage.AsInternal()

		if outInternalMessage.Body == nil {
			continue
		}

		if err := tlb.LoadFromCell(&stonfiPart2, outInternalMessage.Body.BeginParse()); err != nil {
			continue
		}

		if stonfiPart2.OwnerAddr.String() == stonfiPart1.ToAddress.String() {
			part2Found = true
		}
	}

	if !part2Found {
		return nil
	}

	var (
		amountIn  string
		amountOut string
	)

	if stonfiPart2.RefData.Amount0Out.String() == "0" {
		amountIn = fmt.Sprintf("%s %s",
			stonfiPart1.JettonAmount.String(),
			stonfiPart2.RefData.Token0,
		)

		amountOut = fmt.Sprintf("%s %s",
			stonfiPart2.RefData.Amount1Out.String(),
			stonfiPart2.RefData.Token1,
		)
	} else {
		amountIn = fmt.Sprintf("%s %s",
			stonfiPart1.JettonAmount.String(),
			stonfiPart2.RefData.Token1,
		)

		amountOut = fmt.Sprintf("%s %s",
			stonfiPart2.RefData.Amount0Out.String(),
			stonfiPart2.RefData.Token0,
		)
	}

	logrus.Info("[STON.FI] new swap found!")
	logrus.Info("[STON.FI] swaper: ", stonfiPart1.ToAddress)
	logrus.Info("[STON.FI] amount input: ", amountIn)
	logrus.Info("[STON.FI] amount out: ", amountOut)

	return nil
}
