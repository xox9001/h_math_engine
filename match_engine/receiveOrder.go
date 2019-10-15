package match_engine

import (
	"fmt"
	"h-exchange_dev_v0.1/libs/types"
	"log"
	"math/big"
	"strconv"
	"time"
)

//接收http订单数据
func ReceiveLimitOrderByHttp(orders <- chan *htypes.OrderParams){

	//初始化订单，执行撮合
	for {
		orderParams := <- orders

		if _,error := Markets[orderParams.Market]; error == false {
			log.Println("订单数据错误，Market不存在！",orderParams)
			continue
		}

		currentTime := time.Now().Unix()

		order := new(htypes.OrderT)
		order.Id,_ = strconv.ParseInt(orderParams.OrderId,10,64)
		order.Type = htypes.MARKET_ORDER_TYPE_LIMIT
		order.Side,_  = strconv.Atoi(orderParams.Side)
		order.CreatedTime = currentTime
		order.UpdateTime = currentTime
		order.UserId,_ = strconv.ParseInt(orderParams.UserId,10,64)
		order.MarketName = orderParams.Market
		order.Source = orderParams.Source
		//设置精度
		//这里考虑 unsafe 包对地址进行操作

		price,_ := strconv.ParseFloat(orderParams.Price,64)
		amount,_ := strconv.ParseFloat(orderParams.Amount,64)
		tfee,_ := strconv.ParseFloat(orderParams.TakeFee,64)
		mfee,_ := strconv.ParseFloat(orderParams.MakerFee,64)

		order.Price,_ = new(big.Float).SetString(fmt.Sprintf("%.12f",price))
		order.Amount,_ = new(big.Float).SetString(fmt.Sprintf("%.12f",amount))
		order.TakeFee,_ = new(big.Float).SetString(fmt.Sprintf("%.12f",tfee))
		order.MakeFee,_ = new(big.Float).SetString(fmt.Sprintf("%.12f",mfee))
		order.Left = big.NewFloat(0).Copy(order.Amount)
		order.Freeze = big.NewFloat(0)
		order.DealInfo.Stock = big.NewFloat(0)
		order.DealInfo.Money = big.NewFloat(0)
		order.DealInfo.Fee = big.NewFloat(0)


		//推送到市场
		marketData := Markets[orderParams.Market]

		marketData.Lock.Lock()
			//限价单
			if order.Type == htypes.MARKET_ORDER_TYPE_LIMIT {
				if order.Side == htypes.ORDER_SIDE_BID {
					ExecLimitBidOrder(order,marketData)
				}else if order.Side == htypes.ORDER_SIDE_ASK {
					ExecLimitAskOrder(order,marketData)
				}
			}

			if order.Left.Cmp(big.NewFloat(0)) == 0 {
				log.Println("Side: [",order.Side,"] ,Order[",order.Id,"] is Finish")
				//order finish func,free order
				//runtime.GC()
				order = nil
			}else {
				log.Println("order Input list")
				FirstPutOrderToMarket(order,marketData)
			}

		marketData.Lock.Unlock()

	}

}
