package htypes

//全局消息结构，传递不同的消息
type GlobalMessage struct {
	PutOrderChanByHttp chan *OrderParams
}

//全局资源结构，共用资源
type Resource struct {
	RedisPool interface{}
	MysqlPool interface{}
}

var GlobalMsg *GlobalMessage

func init(){
	GlobalMsg = new(GlobalMessage)
	GlobalMsg.PutOrderChanByHttp = make(chan *OrderParams,100000)
}