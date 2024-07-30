package storage

import (
	"time"

	"github.com/shopspring/decimal"
)

type DedustSwap struct {
	Id            uint64 `gorm:"primaryKey;autoIncrement:true;"`
	PoolAddress   string
	AssetIn       string
	AmountIn      decimal.Decimal
	AssetOut      string
	AmountOut     decimal.Decimal
	SenderAddress string
	Reserve0      decimal.Decimal
	Reserve1      decimal.Decimal
	CreatedAt     time.Time
	ProcessedAt   time.Time
}
