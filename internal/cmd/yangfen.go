package cmd

import (
	"fmt"

	yangfenBusiness "github.com/armylong/armylong-go/internal/business/yangfen"
	"github.com/urfave/cli/v2"
)

// 阳粉管理命令
type yangfenCmd struct{}

var YangfenCmd = &yangfenCmd{}

// 阳粉操作入口，按action分发
func (d *yangfenCmd) YangfenHandler(c *cli.Context) error {
	ctx := c.Context

	action := ""
	if c.NArg() > 0 {
		action = c.Args().Get(0)
	}

	uid := c.String("uid")
	amount := c.Int("amount")
	toUid := c.String("to-uid")
	expireSec := c.Int64("expire-sec")
	transactionId := c.String("transaction-id")

	if action == "" {
		fmt.Println("错误: action 不能为空")
		fmt.Println("可用命令: balance, recharge, consume, transfer, refund, transactions, clear")
		return nil
	}

	switch action {
	case "balance":
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return nil
		}
		balance, err := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		if err != nil {
			fmt.Printf("查询余额失败: %v\n", err)
			return nil
		}
		fmt.Printf("用户 %s 当前余额: %d\n", uid, balance)
		return nil

	case "recharge":
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return nil
		}
		if amount <= 0 {
			fmt.Println("错误: amount 必须大于0")
			return nil
		}
		err := yangfenBusiness.YangfenBusiness.Recharge(ctx, uid, amount, expireSec)
		if err != nil {
			fmt.Printf("充值失败: %v\n", err)
			return nil
		}
		balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		fmt.Printf("✓ 充值成功，当前余额: %d\n", balance)
		return nil

	case "consume":
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return nil
		}
		if amount <= 0 {
			fmt.Println("错误: amount 必须大于0")
			return nil
		}
		err := yangfenBusiness.YangfenBusiness.Consume(ctx, uid, amount)
		if err != nil {
			fmt.Printf("消费失败: %v\n", err)
			return nil
		}
		balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		fmt.Printf("✓ 消费成功，当前余额: %d\n", balance)
		return nil

	case "transfer":
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return nil
		}
		if toUid == "" {
			fmt.Println("错误: to-uid 不能为空")
			return nil
		}
		if amount <= 0 {
			fmt.Println("错误: amount 必须大于0")
			return nil
		}
		err := yangfenBusiness.YangfenBusiness.Transfer(ctx, uid, toUid, amount)
		if err != nil {
			fmt.Printf("转账失败: %v\n", err)
			return nil
		}
		fromBalance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		toBalance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, toUid)
		fmt.Printf("✓ 转账成功\n")
		fmt.Printf("  转出账户 %s 余额: %d\n", uid, fromBalance)
		fmt.Printf("  转入账户 %s 余额: %d\n", toUid, toBalance)
		return nil

	case "refund":
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return nil
		}
		if transactionId == "" {
			fmt.Println("错误: transaction-id 不能为空")
			return nil
		}
		err := yangfenBusiness.YangfenBusiness.Refund(ctx, uid, transactionId)
		if err != nil {
			fmt.Printf("退款失败: %v\n", err)
			return nil
		}
		balance, _ := yangfenBusiness.YangfenBusiness.GetBalance(ctx, uid)
		fmt.Printf("✓ 退款成功，当前余额: %d\n", balance)
		return nil

	case "transactions":
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return nil
		}
		transactions, err := yangfenBusiness.YangfenBusiness.GetTransactions(ctx, uid)
		if err != nil {
			fmt.Printf("获取交易记录失败: %v\n", err)
			return nil
		}
		fmt.Printf("用户 %s 交易记录 (共 %d 条):\n", uid, len(transactions))
		fmt.Println("========================================")
		for i, t := range transactions {
			fmt.Printf("[%d] ID: %v\n", i+1, t["id"])
			fmt.Printf("    类型: %v\n", t["type"])
			fmt.Printf("    金额: %v\n", t["amount"])
			fmt.Printf("    余额: %v\n", t["balance"])
			fmt.Printf("    描述: %v\n", t["description"])
			fmt.Println("----------------------------------------")
		}
		return nil

	case "clear":
		if uid == "" {
			fmt.Println("错误: uid 不能为空")
			return nil
		}
		err := yangfenBusiness.YangfenBusiness.ClearData(ctx, uid)
		if err != nil {
			fmt.Printf("清除数据失败: %v\n", err)
			return nil
		}
		fmt.Printf("✓ 用户 %s 数据已清除\n", uid)
		return nil

	default:
		fmt.Printf("未知命令: %s\n", action)
		fmt.Println("可用命令: balance, recharge, consume, transfer, refund, transactions, clear")
	}
	return nil
}
