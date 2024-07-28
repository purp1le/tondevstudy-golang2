package scanner

import (
	"ton-lessons2/internal/structures"

	"github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/tlb"
	"gorm.io/gorm"
)


func (s *scanner) processTransaction(
	trans *tlb.Transaction,
	dbtx *gorm.DB,
) error {
	// 1 - сделать прием тонов и жетонв на кошелек наш, с комментарием
	// 2 - выводить в консоль ВСЕ трансферы жетонов

	if trans.IO.In.MsgType != tlb.MsgTypeInternal {
		return nil
	}

	inTrans := trans.IO.In.AsInternal()

	var transferNotification structures.TransferNotification
	
	if inTrans.Body == nil {
		return nil
	}


	if err := tlb.LoadFromCell(&transferNotification, inTrans.Body.BeginParse()); err != nil {
		return nil
	}

	logrus.Info("[JTN] transfer notification!")
	logrus.Info("[JTN] from: ", transferNotification.Sender)
	logrus.Info("[JTN] to: ", inTrans.DstAddr)
	logrus.Info("[JTN] amount: ", transferNotification.Amount)

	
	return nil
}