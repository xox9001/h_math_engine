package httpserver

import (
	"encoding/json"
	"fmt"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"h-exchange_dev_v0.1/libs/types"
	"h-exchange_dev_v0.1/match_engine"
	"reflect"
	"runtime"
)

var Port string

func init(){
	Port = ":32777"

}

/**
	限价单数据接收
	内部订单推送方法，在此方法接收数据时，订单必须为可进行交易并且资产已经冻结状态
	Method: POST
	Protocol: HTTP
	Request Data: Json Data
	Data Format:
	{
		`json:"user_id"`
		`json:"market"`
		`json:"side"`
		`json:"amount"`
		`json:"price"`
		`json:"take_fee"`
		`json:"make_fee"`
		`json:"source"`
	}
	Response Data:
	{
		"status": [0|1],
		"msg": "",
		"data":[]
	}

 */
func receiveLimitOrder(ctx *routing.Context) error{

	var reqBody = ctx.Request.Body()
	var responseData = new(htypes.ResponseDataFormat)
	var rep []byte
	//默认为失败
	responseData.Msg = "[Error]:Data Is Empty,Please Check."

	//检查是否有数据
	if len(reqBody) > 0 {

		//默认应答为正常
		responseData.Status = 1
		responseData.Msg = "Success"
		orderParams := new(htypes.OrderParams)
		err := json.Unmarshal(reqBody,orderParams)

		if err != nil {
			responseData.Status = 0
			responseData.Msg = "[Error]:Data Format Error,Please Check."
			goto RESPON
		}

		orderParamsV := reflect.ValueOf(orderParams).Elem()
		//检查数据完整性，不能为空
		//TODO: 需要完善数据正确检查，建立校验规则,考虑使用 github: ozzo-validation
		for i:=0;i< orderParamsV.NumField();i++{

			field := orderParamsV.Field(i)
			value := field.String()

			if value == "" {
				responseData.Status = 0
				responseData.Msg = fmt.Sprintf("[Error]: [%s] Data is Empty,Please Check",orderParamsV.Type().Field(i).Name)
				break
			}
		}

		if responseData.Status == 1{
			htypes.GlobalMsg.PutOrderChanByHttp <- orderParams
		}
	}

	goto RESPON


RESPON:
	rep,_ = json.Marshal(responseData)
	//TODO: 公共配置
	ctx.Response.Header.Set("Content-Type","application/json")
	ctx.Write(rep)
	return nil
}

func getAskList(ctx *routing.Context) error{
	str := fmt.Sprintf("卖单数量:%d",match_engine.Markets["CEO-QC"].Asks.Len-1)

	fmt.Fprint(ctx,str)

	return nil
}

func gcAction(ctx *routing.Context) error{
	runtime.GC()


	fmt.Fprint(ctx,"GC Success")
	return nil
}

func getBidList(ctx *routing.Context) error{

	str := fmt.Sprintf("买单数量:%d",match_engine.Markets["CEO-QC"].Bids.Len-1)

	fmt.Fprint(ctx,str)

	return nil
}

func AccessHttpInit(){

	router := routing.New()

	router.Post("/putLimitOrder", receiveLimitOrder)
	router.Get("/asklist", getAskList)
	router.Get("/bidlist", getBidList)
	router.Get("/gc", gcAction)

	serv := &fasthttp.Server{
		Concurrency: 100000,
		Handler:router.HandleRequest,
	}

	panic(serv.ListenAndServe(Port))

}

