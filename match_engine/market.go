package match_engine

import (
	"errors"
	"fmt"
	"h-exchange_dev_v0.1/libs/skiplist"
	"h-exchange_dev_v0.1/libs/types"
	"log"
	"math/big"
	"sync"
	"time"
)

var (
	MarketConfigs map[string]*htypes.MarketConfig
	Markets map[string]*htypes.MarketT
	DealsId int64
	)

//初始化市场数据
func InitMarket(){

	MarketConfigs = loadMarkets()
	Markets = make(map[string]*htypes.MarketT)

	if MarketConfigs != nil && len(MarketConfigs) > 0 {

		for name,marketconfig := range MarketConfigs {

			if _,ok := Markets[name]; !ok {
				Markets[name] = CreateMarket(marketconfig)
			}
		}
	}

	count := len(Markets)

	if count > 0 {
		log.Println("[Success]: Market初始化成功,共初始化[",count,"] 个Market。Demo Market:CEO-QC")
	}else {
		log.Fatal("[ERROR]:market无数据")
	}
}

/**
	初始化市场名称，费率精度，交易
 */
func loadMarkets() map[string]*htypes.MarketConfig{

	markets := make(map[string]*htypes.MarketConfig)

	//demo CEO/QC
	//config by json
	name := "CEO-QC"
	config := new(htypes.MarketConfig)
	config.Name = name
	//市场费率精度
	config.FeePrec = 4
	//最小交易量
	config.MinAccount = big.NewFloat(0.001)
	config.StockName = "CEO"
	config.StockPrec = 8
	config.MoneyName = "QC"
	config.MoneyPrec = 8
	markets[name] = config

	return markets
}

func CreateMarket(config *htypes.MarketConfig) *htypes.MarketT{

	market := new(htypes.MarketT)
	market.Name = config.Name
	market.Stock = config.StockName
	market.Money = config.MoneyName
	market.StockPrec = config.StockPrec
	market.MoneyPrec = config.MoneyPrec
	market.FeePrec = config.FeePrec
	market.MinAmount = config.MinAccount
	market.Orders = new(sync.Map)
	market.Users = new(sync.Map)
	market.Asks = skiplist.NewSkiplist(&htypes.OrderListType{})
	market.Bids = skiplist.NewSkiplist(&htypes.OrderListType{})

	return market
}

/**
卖单撮合
这里可以考虑如何减少内存gc操作，将买卖撮合持久化，独立为goroutine
 */
func ExecLimitAskOrder(taker *htypes.OrderT,market *htypes.MarketT)(bool,error){
	fmt.Println("卖单撮合")

	var price = new(big.Float)
	var amount = new(big.Float)
	var deal = new(big.Float)
	var ask_fee = new(big.Float)
	var bid_fee = new(big.Float)
	var node *skiplist.SkiplistNode
	var ite = market.Bids.SkiplistGetIter()
	var zero = big.NewFloat(0)

	for {

		node = ite.SkiplistNext()

		if node == nil {
			break
		}

		if taker.Left.Cmp(zero) == 0 {
			break;
		}

		maker := node.Value.(*htypes.OrderT)

		if taker.Price.Cmp(maker.Price) > 0{
			break
		}

		*price = *maker.Price

		if taker.Left.Cmp(maker.Left) < 0 {
			amount.Set(taker.Left)
		}else {
			amount.Set(maker.Left)
		}

		deal = deal.Mul(price,amount)
		ask_fee = ask_fee.Mul(deal,taker.TakeFee)
		bid_fee = bid_fee.Mul(amount,maker.MakeFee)

		curTime := time.Now().Unix()
		//========= Taker Action Lock
		taker.Lock.Lock()

		taker.UpdateTime = curTime
		taker.Left.Sub(taker.Left,amount)
		taker.DealInfo.Stock.Add(taker.DealInfo.Stock,amount)
		taker.DealInfo.Money.Add(taker.DealInfo.Money,deal)
		taker.DealInfo.Fee.Add(taker.DealInfo.Fee,ask_fee) //taker吃买单,承担卖单手续费

		taker.Lock.Unlock()
		//========= Taker Action UnLock
		//Push Msg
		//TODO: PUSH Deal Histiory MSG
		//deal_id := atomic.AddInt64(&DealsId,1)
		//TODO: end
		/**
		Taker Action:
			1、生成并推送交易历史消息（实时消息）
			1.1、插入订单交易历史记录 （实时消息）
			2、减少卖家(taker)挂单卖的可用资产（stock）
			3、推送资产减少消息  （实时消息）
			4、卖家(taker)的可用资产增加 (money)
			5、推送资产增加交易历史记录 （实时消息）
			6、推送实时成交消息（实时消息）
			7、减少卖家(taker)的卖单手续费（sub money ask_fee）
			8、推送手续费交易消息 （实时消息）
			TIPS: 每一次消息推送都有数据持久化操作
		 */

		//========= maker Action Lock
		maker.Lock.Lock()
		maker.UpdateTime = curTime
		maker.Left.Sub(maker.Left,amount)
		maker.Freeze.Sub(maker.Freeze,deal)
		maker.DealInfo.Stock.Add(maker.DealInfo.Stock,amount)
		maker.DealInfo.Money.Add(maker.DealInfo.Money,deal)
		maker.DealInfo.Fee.Add(maker.DealInfo.Money,bid_fee)

		/**
			Maker Action:
			1、减少买家冻结资金 (money)
			2、推送买家与市场成交消息
			3、记录用户交易历史、资金变更历史
			4、增加买家买入的资产 (stock)
			5、增加资产历史记录
			6、买家手续费处理，减少可用的买入资产(stock)
			7、推送买家资产变更历史记录

		 */

		maker.Lock.Unlock()
		//========= maker Action ULock

		fmt.Println("卖单成交：")
		fmt.Println("成交价:",price.Text('f',12))
		fmt.Println("成交量:",amount.Text('f',12))
		fmt.Println("交易额:",deal.Text('f',12))
		fmt.Println("买单费率:",bid_fee.Text('f',12))
		fmt.Println("卖单费率:",ask_fee.Text('f',12))

		if maker.Left.Cmp(zero) == 0{
			log.Println("Side:[",maker.Side,"],order [",maker.Id,"] 已经完成")
			market.Bids.DeleteNode(node)
		}

		/**
			如果maker.left 为 0 ，则触发订单完成操作并推送消息，否则推送订单更新消息
			如果买单剩余数量为0，则推送订单完成消息
			触发订单完成事件，将完成订单清理出订单薄
			如果订单还有余量，则推送订单更新消息
		 */

	}

	price,amount,deal,ask_fee,bid_fee,node,ite = nil,nil,nil,nil,nil,nil,nil
	zero = nil

	return true,nil
}

/**
撮合限价买单
这里可以考虑如何减少内存gc操作，将买卖撮合持久化，独立为goroutine
 */
func ExecLimitBidOrder(order *htypes.OrderT,market *htypes.MarketT)(bool,error){

	fmt.Println("买单撮合")

	price,amount,deal,ask_fee,bid_fee := new(big.Float),new(big.Float),new(big.Float),new(big.Float),new(big.Float)
	zero := big.NewFloat(0)
	var ite = market.Asks.SkiplistGetIter()

	for {

		var node = ite.SkiplistNext()

		if node == nil {
			break
		}

		if order.Left.Cmp(zero) == 0 {
			break;
		}

		maker := node.Value.(*htypes.OrderT)

		if order.Price.Cmp(maker.Price) < 0 {
			break
		}

		*price = *maker.Price

		if order.Left.Cmp(maker.Left) < 0 {
			amount.Set(order.Left)
		}else{
			amount.Set(maker.Left)
		}

		deal.Mul(price,amount)
		ask_fee.Mul(deal,maker.MakeFee)
		bid_fee.Mul(amount,order.TakeFee)
		curTime := time.Now().Unix()

		// ====== Taker Action:

			order.Lock.Lock()

					order.UpdateTime = curTime
					order.Left.Sub(order.Left,amount)
					order.DealInfo.Stock.Add(order.DealInfo.Stock,amount)
					order.DealInfo.Money.Add(order.DealInfo.Money,deal)
					order.DealInfo.Fee.Add(order.DealInfo.Fee,bid_fee)
			order.Lock.Unlock()

		/**
		  	Taker Step :
				1、增加订单成交历史记录和实时消息
				2、增加买单用户资产变更历史（减少用户可用余额，也就是减少冻结了的订单金额）
				3、增加买单用户资产变更历史（增加用户已买到的资产变更历史记录）
				4、增加买单用户交易手续费历史（增加用户已买到资产的手续费历史记录）
			Balance Step:
				1、买单用户可用资金减少（money）
				2、买单用户可用资产增加（stock）
				3、买单用户已买到资产减去买单手续费(stock)

		 */

		// ====== Taker Action End

		// ====== Maker Action:

			maker.Lock.Lock()

				maker.Left.Sub(maker.Left,amount)
				maker.Freeze.Sub(maker.Freeze,amount)
				maker.DealInfo.Stock.Add(maker.DealInfo.Stock,amount)
				maker.DealInfo.Money.Add(maker.DealInfo.Money,deal)
				maker.DealInfo.Fee.Add(maker.DealInfo.Fee,ask_fee)

			maker.Lock.Unlock()

		/**
		  Maker Step :
			1、增加订单成交历史记录和实时消息
			2、增加买单用户资产变更历史（减少用户可用余额，也就是减少冻结了的订单金额）
			3、增加买单用户资产变更历史（增加用户已买到的资产变更历史记录）
			4、增加买单用户交易手续费历史（增加用户已买到资产的手续费历史记录）
		Balance Step:
			1、买单用户可用资金减少（money）
			2、买单用户可用资产增加（stock）
			3、买单用户已买到资产减去买单手续费(stock)

	 */

		// ====== Maker Action End

		if maker.Left.Cmp(zero) == 0{
			fmt.Println("Side:[",maker.Side,"],order [",maker.Id,"] 已经完成")
			market.Asks.DeleteNode(node)
		}

		/**
					如果maker.left 为 0 ，则触发订单完成操作并推送消息，否则推送订单更新消息
					如果买单剩余数量为0，则推送订单完成消息
					触发订单完成事件，将完成订单清理出订单薄
					如果订单还有余量，则推送订单更新消息
		*/

		fmt.Println("买单成交：")
		fmt.Println("成交价:",price.Text('f',12))
		fmt.Println("成交量:",amount.Text('f',12))
		fmt.Println("交易额:",deal.Text('f',12))
		fmt.Println("买单费率:",bid_fee.Text('f',12))
		fmt.Println("卖单费率:",ask_fee.Text('f',12))

	}

	price,amount,deal,ask_fee,bid_fee,ite = nil,nil,nil,nil,nil,nil
	zero = nil

	return true,nil
}

//将订单插入订单薄
func PutOrderToList(order *htypes.OrderT){



}

//将订单推送到市场中的order与user 列表
func FirstPutOrderToMarket(order *htypes.OrderT,market *htypes.MarketT) (bool,error){

	var msg string = ""
	//input order list
	//if _,ok := market.Orders.Load(order.Id); ok == false{
	//	market.Orders.Store(order.Id,order)
	//}else{
	//	msg = fmt.Sprintf("[Error]:Market Orders 订单重复，订单号：[%d] \n" ,order.Id)
	//	return false,errors.New(msg)
	//}

	//input user order list
	//if userOrders,ok := market.Users.Load(order.UserId); ok == false{
	//	userOrders = &htypes.UserOrders{
	//		UserId:order.UserId,
	//		OrderList: new(sync.Map),
	//	}
	//
	//	userOrders.(*htypes.UserOrders).OrderList.Store(order.Id,order)
	//
	//}else{
	//
	//	orderlist := userOrders.(*htypes.UserOrders).OrderList
	//
	//	if _,ok := orderlist.Load(order.Id);ok == false{
	//		orderlist.Store(order.Id,order)
	//	}else{
	//		msg = fmt.Sprintf("[Error]:Market Order 正常但 UserOrders 订单重复，订单号：[%d],请检查！ \n" ,order.Id)
	//		return false,errors.New(msg)
	//	}
	//
	//}

	//input ask or bid list
	if order.Side == htypes.ORDER_SIDE_ASK {
		market.Asks.InsertNode(order)
		//处理冻结资产 stock
	}else{
		market.Bids.InsertNode(order)
		//处理冻结资产 money
	}

	return true,errors.New(msg)

}