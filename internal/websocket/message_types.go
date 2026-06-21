package websocket

// WebSocket推送消息类型常量
// 业务层通过 libWs.NewMessage(TypeXxx, data) 构造消息，调用 Manager.PushToUsers(uids, msg) 推送

// 拼拼坐
const (
	TypeOrderMatched      = "order_matched"       // 订单匹配成功，司机接单
	TypeOrderCancelled    = "order_cancelled"      // 订单被取消
	TypeTripStatusChanged = "trip_status_changed"  // 行程状态变更（出发/到达/完成等）
	TypeNewOrder          = "new_order"            // 司机端收到新匹配订单
)

// 龙聊消息类型常量
const (
	TypeLongChat = "long_chat"
)