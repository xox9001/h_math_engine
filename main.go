package main

import (
	"h-exchange_dev_v0.1/httpserver"
	"h-exchange_dev_v0.1/libs/types"
	_ "h-exchange_dev_v0.1/libs/utils"
	"h-exchange_dev_v0.1/match_engine"
	"log"
	"runtime"
)


func main()  {
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Println("撮合引擎正在启动")
	go match_engine.ReceiveLimitOrderByHttp(htypes.GlobalMsg.PutOrderChanByHttp)
	match_engine.InitMarket()
	httpserver.AccessHttpInit()

}
