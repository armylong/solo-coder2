package yangfen

import (
	"errors"
	"fmt"

	yangfenBusiness "github.com/armylong/armylong-go/internal/business/yangfen"
	"github.com/armylong/armylong-go/internal/middlewares"
	yangfenCs "github.com/armylong/armylong-go/internal/cs/yangfen"
	"github.com/gin-gonic/gin"
)

// YangfenController 氧分系统
type YangfenController struct{}

// 从上下文获取登录用户ID
func (c *YangfenController) getUid(ctx *gin.Context) (string, error) {
	uid := middlewares.GetLoginUID(ctx)
	if uid == 0 {
		return "", errors.New("请先登录")
	}
	return fmt.Sprintf("%d", uid), nil
}

// 查余额
func (c *YangfenController) ActionGetBalance(ctx *gin.Context, req *yangfenCs.BaseRequest) (*yangfenCs.BalanceResponse, error) {
	uid, err := c.getUid(ctx)
	if err != nil {
		return nil, err
	}

	balance, err := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &yangfenCs.BalanceResponse{
		Uid:     uid,
		Balance: balance,
	}, nil
}

// 充值
func (c *YangfenController) ActionRecharge(ctx *gin.Context, req *yangfenCs.RechargeRequest) (*yangfenCs.BalanceResponse, error) {
	uid, err := c.getUid(ctx)
	if err != nil {
		return nil, err
	}

	if req.Amount <= 0 {
		return nil, errors.New("充值金额必须大于0")
	}

	err = yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, req.Amount, req.ExpireSec)
	if err != nil {
		return nil, err
	}

	balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)

	return &yangfenCs.BalanceResponse{
		Uid:     uid,
		Balance: balance,
	}, nil
}

// 消费
func (c *YangfenController) ActionConsume(ctx *gin.Context, req *yangfenCs.ConsumeRequest) (*yangfenCs.BalanceResponse, error) {
	uid, err := c.getUid(ctx)
	if err != nil {
		return nil, err
	}

	if req.Amount <= 0 {
		return nil, errors.New("消费金额必须大于0")
	}

	err = yangfenBusiness.YangfenBusiness.Consume(ctx, uid, req.Amount)
	if err != nil {
		return nil, err
	}

	balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)

	return &yangfenCs.BalanceResponse{
		Uid:     uid,
		Balance: balance,
	}, nil
}

// 转账
func (c *YangfenController) ActionTransfer(ctx *gin.Context, req *yangfenCs.TransferRequest) (*yangfenCs.TransferResponse, error) {
	uid, err := c.getUid(ctx)
	if err != nil {
		return nil, err
	}

	if req.ToUid == "" {
		return nil, errors.New("目标用户不能为空")
	}

	if req.Amount <= 0 {
		return nil, errors.New("转账金额必须大于0")
	}

	err = yangfenBusiness.YangfenBusiness.Transfer(ctx, uid, req.ToUid, req.Amount)
	if err != nil {
		return nil, err
	}

	fromBalance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
	toBalance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, req.ToUid)

	return &yangfenCs.TransferResponse{
		FromUid:     uid,
		FromBalance: fromBalance,
		ToUid:       req.ToUid,
		ToBalance:   toBalance,
	}, nil
}

// 退款
func (c *YangfenController) ActionRefund(ctx *gin.Context, req *yangfenCs.RefundRequest) (*yangfenCs.BalanceResponse, error) {
	uid, err := c.getUid(ctx)
	if err != nil {
		return nil, err
	}

	if req.TransactionId == "" {
		return nil, errors.New("交易号不能为空")
	}

	err = yangfenBusiness.YangfenBusiness.Refund(ctx, uid, req.TransactionId)
	if err != nil {
		return nil, err
	}

	balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)

	return &yangfenCs.BalanceResponse{
		Uid:     uid,
		Balance: balance,
	}, nil
}

// 查交易列表
func (c *YangfenController) ActionGetTransactions(ctx *gin.Context, req *yangfenCs.BaseRequest) (*yangfenCs.TransactionListResponse, error) {
	uid, err := c.getUid(ctx)
	if err != nil {
		return nil, err
	}

	transactions, err := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &yangfenCs.TransactionListResponse{
		List:  convertTransactions(transactions),
		Total: len(transactions),
	}, nil
}

// 交易记录转响应结构
func convertTransactions(transactions []map[string]any) []yangfenCs.TransactionRecord {
	result := make([]yangfenCs.TransactionRecord, 0, len(transactions))
	for _, t := range transactions {
		record := yangfenCs.TransactionRecord{}
		if id, ok := t["id"].(string); ok {
			record.Id = id
		}
		if uid, ok := t["uid"].(string); ok {
			record.Uid = uid
		}
		if txType, ok := t["type"].(string); ok {
			record.Type = txType
		}
		if amount, ok := t["amount"].(int); ok {
			record.Amount = amount
		}
		if balance, ok := t["balance"].(int); ok {
			record.Balance = balance
		}
		if desc, ok := t["description"].(string); ok {
			record.Description = desc
		}
		if createdAt, ok := t["createdAt"].(float64); ok {
			record.CreatedAt = int64(createdAt)
		}
		result = append(result, record)
	}
	return result
}
