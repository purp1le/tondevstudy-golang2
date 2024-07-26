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
)

type scanner struct {
	api            *ton.APIClient
	lastBlock      storage.Block
	shardLastSeqno map[string]uint32
}

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
	}
	newShards = append(newShards, currentShards...)

	var txList []*tlb.Transaction

	var wg sync.WaitGroup
	var tombGetTransactions tomb.Tomb
	allDone := make(chan struct{})
	for _, shard := range newShards {
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
					for i := 0; i<3 || err != nil; i++ {
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
	for _, transaction := range txList {
		if err := s.processTransaction(transaction); err != nil {
			return err
		}
	}

	if err := s.addBlock(*master); err != nil {
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
