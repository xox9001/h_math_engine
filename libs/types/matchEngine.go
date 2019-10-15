package htypes

import (
	"h-exchange_dev_v0.1/libs/skiplist"
	"math/big"
	"sync"
)

const (
	MARKET_ORDER_TYPE_LIMIT = 1
	MARKET_ORDER_TYPE_MARKET = 2
	ORDER_SIDE_ASK = 1
	ORDER_SIDE_BID = 2
)

type OrderT struct {
	Lock sync.Mutex
	Id int64
	Type int
	Side int
	CreatedTime int64
	UpdateTime int64
	UserId  int64
	MarketName string
	Source  string
	Price 	*big.Float
	Amount  *big.Float
	TakeFee  *big.Float
	MakeFee  *big.Float
	Left	*big.Float
	Freeze	*big.Float
	DealInfo
}

type DealInfo struct {
	Stock 	*big.Float
	Money	*big.Float
	Fee		*big.Float
}

type MarketT struct {
	Lock sync.Mutex
	Name string
	Stock string
	Money string
	StockPrec int
	MoneyPrec int
	FeePrec int
	MinAmount *big.Float
	Orders *sync.Map
	Users  *sync.Map
	Asks   *skiplist.Skiplist
	Bids   *skiplist.Skiplist
}

type UserOrders struct {
	UserId int64
	OrderList *sync.Map
}


type Asset struct {
	Name string
	SavePrec int
	ShowPrec int
}

//market配置单元
type MarketConfig struct{
	Name string
	FeePrec int
	MinAccount *big.Float
	StockName string
	StockPrec int
	MoneyName string
	MoneyPrec int
}

/**
	订单类型方法
	SkiplistType interface
	@path: libs/skiplist/types.go
 */

type OrderListType struct {}

func (_ *OrderListType)Compare(value interface{},value1 interface{}) (int){

	//TODO: 考虑数据异常
	order1,_ := value.(*OrderT)
	order2,_ := value1.(*OrderT)

	if order1.Id == order2.Id {
		return 0
	}

	if order1.Type != order2.Type {
		return 1
	}

	var cmp int

	if order1.Side == ORDER_SIDE_ASK {
		cmp = order1.Price.Cmp(order2.Price)
	}else{
		cmp = order2.Price.Cmp(order1.Price)
	}

	if cmp != 0 {
		return cmp
	}

	if order1.Id > order2.Id {
		return 1
	}else{
		return -1
	}
}

func (_ *OrderListType)Dup(value interface{}) (interface{},bool){

	return nil,false
}

func (_ *OrderListType)Free(value interface{}) (bool){

	return false
}