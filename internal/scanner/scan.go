package scanner

import (
	"context"
	"sync"
	"time"
	"ton-lessons2/internal/app"
	"ton-lessons2/internal/storage"

	"github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"gopkg.in/tomb.v1"
	"gorm.io/gorm"
)

type scanner struct {
	api            *ton.APIClient
	lastBlock      storage.Block
	shardLastSeqno map[string]uint32
}

// * - pointer
// & - link
// &
func NewScanner() (*scanner, error) {
	client := liteclient.NewConnectionPool()

	if err := client.AddConnectionsFromConfig(
		context.Background(),
		app.CFG.MainnetConfig,
	); err != nil {
		return nil, err
	}

	api := ton.NewAPIClient(client)

	return &scanner{
		api:            api,
		lastBlock:      storage.Block{},
		shardLastSeqno: make(map[string]uint32),
	}, nil
}

func (s *scanner) Listen() {
	logrus.Info("[SCN] start scanning blocks")

	if err := app.DB.Last(&s.lastBlock).Error; err != nil {
		s.lastBlock.SeqNo = 0
	} else {
		s.lastBlock.SeqNo = 0
	}

	if s.lastBlock.SeqNo == 0 {
		lastMaster, err := s.api.GetMasterchainInfo(context.Background())
		for err != nil {
			time.Sleep(time.Second)
			logrus.Error("[SCN] error when get last master: ", err)
			lastMaster, err = s.api.GetMasterchainInfo(context.Background())
		}

		s.lastBlock.SeqNo = lastMaster.SeqNo
		s.lastBlock.Shard = lastMaster.Shard
		s.lastBlock.WorkChain = lastMaster.Workchain
	}

	masterBlock, err := s.api.LookupBlock(
		context.Background(),
		s.lastBlock.WorkChain,
		s.lastBlock.Shard,
		s.lastBlock.SeqNo,
	)
	for err != nil {
		time.Sleep(time.Second)
		logrus.Error("[SCN] error when lookup block: ", err)
		masterBlock, err = s.api.LookupBlock(
			context.Background(),
			s.lastBlock.WorkChain,
			s.lastBlock.Shard,
			s.lastBlock.SeqNo,
		)
	}

	firstShards, err := s.api.GetBlockShardsInfo(
		context.Background(),
		masterBlock,
	)
	for err != nil {
		time.Sleep(time.Second)
		logrus.Error("[SCN] error when get first shards: ", err)
		firstShards, err = s.api.GetBlockShardsInfo(
			context.Background(),
			masterBlock,
		)
	}

	for _, shard := range firstShards {
		s.shardLastSeqno[s.getShardID(shard)] = shard.SeqNo
	}

	s.processBlocks()
}

func (s *scanner) processBlocks() {
	for {
		masterBlock, err := s.api.LookupBlock(
			context.Background(),
			s.lastBlock.WorkChain,
			s.lastBlock.Shard,
			s.lastBlock.SeqNo,
		)
		for err != nil {
			time.Sleep(time.Second)
			logrus.Error("[SCN] error when lookup block: ", err)
			masterBlock, err = s.api.LookupBlock(
				context.Background(),
				s.lastBlock.WorkChain,
				s.lastBlock.Shard,
				s.lastBlock.SeqNo,
			)
		}

		scanErr := s.processMcBlock(masterBlock)
		for scanErr != nil {
			logrus.Error("[SCN] mc block err: ", err)
			time.Sleep(time.Second * 2)
			scanErr = s.processMcBlock(masterBlock)
		}
	}
}

// jetton transfer
// a wallet -> a jetton wallet -> b jetton wallet -> b wallet(notification)
//								  				-> a wallet(excesses)

// struct block
// 32bit - seqno
// 32bit - workchain
// 64bit - shard
// 32bit - time

// func a() -> create struct block
// func b(block struct.block) struct.block -> change struct block
// newBlock = b(newBlock)
// allocate memory for new struct
// -> copy struct
// change struct
// return struct.block (copy)
// func c(block *struct.block) -> change struct block
// c(&newBlock)
// uint32 -> 0xa83127
// block.seqno = 123
// nil - default pointer value

// input message 1
// output messages 2



func (s *scanner) processMcBlock(master *ton.BlockIDExt) error {
	timeStart := time.Now()

	currentShards, err := s.api.GetBlockShardsInfo(
		context.Background(),
		master,
	)
	if err != nil {
		return err
	}

	if len(currentShards) == 0 {
		logrus.Debugf("[SCN] block [%d] without shards", master.SeqNo)
		return nil
	}

	var newShards []*ton.BlockIDExt

	for _, shard := range currentShards {
		notSeen, err := s.getNonSeenShards(context.Background(), shard)
		if err != nil {
			return err
		}

		s.shardLastSeqno[s.getShardID(shard)] = shard.SeqNo
		newShards = append(newShards, notSeen...)
	}

	if len(newShards) == 0 {
		newShards = currentShards
	} else {
		newShards = append(newShards, currentShards...)
	}

	if len(newShards) == 0 {
		return nil
	}

	var txList []*tlb.Transaction


	uniqueShards := s.getUniqueShards(newShards)
	var wg sync.WaitGroup
	var tombGetTransactions tomb.Tomb
	allDone := make(chan struct{})
	for _, shard := range uniqueShards {
		var (
			fetchedIDs []ton.TransactionShortInfo
			after      *ton.TransactionID3
			more       = true
		)

		for more {
			fetchedIDs, more, err = s.api.GetBlockTransactionsV2(
				context.Background(),
				shard,
				100,
				after,
			)
			if err != nil {
				return err
			}

			if more {
				after = fetchedIDs[len(fetchedIDs)-1].ID3()
			}

			for _, id := range fetchedIDs {
				wg.Add(1)
				go func(shard *tlb.BlockInfo, account []byte, lt uint64) {
					defer wg.Done()
					tx, err := s.api.GetTransaction(
						context.Background(),
						shard,
						address.NewAddress(0, 0, account),
						lt,
					)
					for i := 0; i < 3 || err != nil; i++ {
						time.Sleep(time.Second)
						tx, err = s.api.GetTransaction(
							context.Background(),
							shard,
							address.NewAddress(0, 0, account),
							lt,
						)
					}
					if err != nil {
						tombGetTransactions.Kill(err)
					}
					txList = append(txList, tx)
				}(shard, id.Account, id.LT)

			}
		}

	}

	go func() {
		wg.Wait()
		close(allDone)
	}()

	select {
	case <-allDone:
	case <-tombGetTransactions.Dying():
		logrus.Error("[SCN] err when get transactions: ", tombGetTransactions.Err())
		return tombGetTransactions.Err()
	}
	tombGetTransactions.Done()

	// process transactions

	dbtx := app.DB.Begin()

	var wgTrans sync.WaitGroup
	allDoneTrans := make(chan struct{})
	var tombTrans tomb.Tomb

	for _, transaction := range txList {
		wg.Add(1)

		go func(dbtx *gorm.DB, transaction *tlb.Transaction) {
			defer wg.Done()
			if err := s.processTransaction(transaction, dbtx); err != nil {
				tombTrans.Kill(err)
			}
		}(dbtx, transaction)
	}
	

	go func() {
		wgTrans.Wait()
		close(allDoneTrans)
	}()

	select {
	case <-allDoneTrans:
	case <-tombTrans.Dying():
		logrus.Error("[SCN] err when process transactions: ", err)
		dbtx.Rollback()
		return tombTrans.Err()
	}
	tombTrans.Done()

	if err := s.addBlock(*master, dbtx); err != nil {
		dbtx.Rollback()
		return err
	}

	if err := dbtx.Commit().Error; err != nil {
		logrus.Error("[SCN] dbtx commit err: ", err)
		return err
	}

	lastSeqno, err := s.getLastBlockSeqno()
	if err != nil {
		logrus.Infof("[SCN] success process block [%d] time to process block [%0.2fs] trans count [%d]",
			master.SeqNo,
			time.Since(timeStart).Seconds(),
			len(txList),
		)
	} else {
		logrus.Infof("[SCN] success process block [%d|%d] time to process block [%0.2fs] trans count [%d]",
			master.SeqNo,
			lastSeqno,
			time.Since(timeStart).Seconds(),
			len(txList),
		)
	}
	return nil
}
