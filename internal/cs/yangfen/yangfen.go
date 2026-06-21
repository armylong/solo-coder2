package yangfen

// 基础请求
type BaseRequest struct {
	Uid string `json:"uid" form:"uid"` // 用户ID
}

// 余额-响应
type BalanceResponse struct {
	Uid     string `json:"uid"`     // 用户ID
	Balance int    `json:"balance"` // 余额
}

// 充值-请求
type RechargeRequest struct {
	BaseRequest
	Amount    int   `json:"amount" form:"amount"`       // 充值金额
	ExpireSec int64 `json:"expire_sec" form:"expire_sec"` // 过期秒数
}

// 消费-请求
type ConsumeRequest struct {
	BaseRequest
	Amount int `json:"amount" form:"amount"` // 消费金额
}

// 转账-请求
type TransferRequest struct {
	BaseRequest
	ToUid  string `json:"toUid" form:"toUid"`   // 目标用户ID
	Amount int    `json:"amount" form:"amount"` // 转账金额
}

// 转账-响应
type TransferResponse struct {
	FromUid     string `json:"from_uid"`     // 转出用户ID
	FromBalance int    `json:"from_balance"` // 转出后余额
	ToUid       string `json:"to_uid"`       // 转入用户ID
	ToBalance   int    `json:"to_balance"`   // 转入后余额
}

// 退款-请求
type RefundRequest struct {
	BaseRequest
	TransactionId string `json:"transactionId" form:"transactionId"` // 交易号
}

// 交易记录
type TransactionRecord struct {
	Id          string `json:"id"`          // 交易号
	Uid         string `json:"uid"`         // 用户ID
	Type        string `json:"type"`        // 交易类型
	Amount      int    `json:"amount"`      // 金额
	Balance     int    `json:"balance"`     // 交易后余额
	Description string `json:"description"` // 描述
	CreatedAt   int64  `json:"createdAt"`   // 创建时间
}

// 交易列表-响应
type TransactionListResponse struct {
	List  []TransactionRecord `json:"list"`  // 交易记录列表
	Total int                 `json:"total"` // 总数
}
