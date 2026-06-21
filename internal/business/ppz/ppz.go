package ppz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	ppzCs "github.com/armylong/armylong-go/internal/cs/ppz"
	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
)

type ppzBusiness struct{}

var PpzBusiness = &ppzBusiness{}

// 检查司机是否被封禁
func (b *ppzBusiness) checkDriverStatus(ctx context.Context, uid int64) error {
	isDriverBanned, err := ppzModel.TbPpzUserModel.IsDriverBanned(uid)
	if err != nil {
		return fmt.Errorf("检查用户状态失败: %w", err)
	}
	if isDriverBanned {
		return errors.New("您的司机状态异常, 不能管理车辆, 如有需要请联系客服")
	}
	return nil
}

// 获取用户信息
func (b *ppzBusiness) GetPpzUserInfo(ctx context.Context, uid int64) (*ppzCs.GetPpzUserInfoResponse, error) {
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	user, err := ppzModel.TbPpzUserModel.GetOrCreateByUid(uid)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	return &ppzCs.GetPpzUserInfoResponse{
		CarCount:     user.CarCount,
		Status:       user.Status,
		DriverStatus: user.DriverStatus,
	}, nil
}

// 添加车辆
func (b *ppzBusiness) AddMyCar(ctx context.Context, uid int64, req *ppzCs.AddMyCarRequest) (*ppzCs.AddMyCarResponse, error) {
	return PpzAuditBusiness.AddMyCar(ctx, uid, req)
}

// 获取我的车辆列表
func (b *ppzBusiness) GetMyCars(ctx context.Context, uid int64, req *ppzCs.GetMyCarsRequest) (*ppzCs.GetMyCarsResponse, error) {
	return PpzAuditBusiness.GetMyCars(ctx, uid, req)
}

// 删除车辆
func (b *ppzBusiness) DeleteMyCar(ctx context.Context, uid int64, req *ppzCs.DeleteMyCarRequest) (*ppzCs.DeleteMyCarResponse, error) {
	return PpzAuditBusiness.DeleteMyCar(ctx, uid, req)
}

// 检查是否认证司机
func (b *ppzBusiness) CheckDriver(ctx context.Context, uid int64) (*ppzCs.CheckDriverResponse, error) {
	return PpzAuditBusiness.CheckDriver(ctx, uid)
}

// 编辑车辆
func (b *ppzBusiness) EditMyCar(ctx context.Context, uid int64, req *ppzCs.EditMyCarRequest) (*ppzCs.EditMyCarResponse, error) {
	return PpzAuditBusiness.EditMyCar(ctx, uid, req)
}

// 获取车辆详情
func (b *ppzBusiness) GetMyCarDetail(ctx context.Context, uid int64, req *ppzCs.GetCarDetailRequest) (*ppzCs.GetCarDetailResponse, error) {
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	if err := b.checkDriverStatus(ctx, uid); err != nil {
		return nil, err
	}

	var audit *ppzModel.TbPpzCarAudit
	var err error

	if req.AuditId > 0 {
		audit, err = ppzModel.TbPpzCarAuditModel.GetByUidAndId(uid, req.AuditId)
		if err != nil || audit == nil {
			return nil, errors.New("车辆不存在或无权限查看")
		}
	} else if req.CarId > 0 {
		audit, err = ppzModel.TbPpzCarAuditModel.GetByUidAndCarId(uid, req.CarId)
		if err != nil || audit == nil {
			car, err := ppzModel.TbPpzCarsModel.GetByUidAndId(uid, req.CarId)
			if err != nil || car == nil {
				return nil, fmt.Errorf("车辆不存在: %w", err)
			}
			return &ppzCs.GetCarDetailResponse{
				Car: &ppzCs.CarAuditDetail{
					AuditId:            0,
					CarId:              car.CarId,
					Uid:                car.Uid,
					CarModel:           car.CarModel,
					CarLicensePhoto:    car.CarLicensePhoto,
					DriverLicensePhoto: car.DriverLicensePhoto,
					LicensePlate:       car.LicensePlate,
					CarColor:           car.CarColor,
					Seats:              car.Seats,
					CarPhoto:           car.CarPhoto,
					Description:        car.Description,
					AuditStatus:        2,
					ReviewStatus:       2,
					AuditData:          nil,
					CreatedAt:          car.CreatedAt,
					UpdatedAt:          car.UpdatedAt,
				},
			}, nil
		}
	} else {
		return nil, errors.New("请提供车辆ID或审核ID")
	}

	carDetail := &ppzCs.CarAuditDetail{
		AuditId:            audit.AuditId,
		CarId:              audit.CarId,
		Uid:                audit.Uid,
		CarModel:           audit.AuditData.CarModel,
		CarLicensePhoto:    audit.AuditData.CarLicensePhoto,
		DriverLicensePhoto: audit.AuditData.DriverLicensePhoto,
		LicensePlate:       audit.AuditData.LicensePlate,
		CarColor:           audit.AuditData.CarColor,
		Seats:              audit.AuditData.Seats,
		CarPhoto:           audit.AuditData.CarPhoto,
		Description:        audit.AuditData.Description,
		AuditStatus:        audit.AuditStatus,
		ReviewStatus:       audit.AuditStatus,
		AuditReason:        audit.AuditReason,
		AuditData:          audit.AuditData,
		CreatedAt:          audit.CreatedAt,
		UpdatedAt:          audit.UpdatedAt,
	}

	return &ppzCs.GetCarDetailResponse{
		Car: carDetail,
	}, nil
}

// 获取地址选择器数据（最近出发地/目的地 + 常用地址）
func (b *ppzBusiness) GetAddressPickerData(ctx context.Context, uid int64) (*ppzCs.AddressPickerDataResponse, error) {
	resp := &ppzCs.AddressPickerDataResponse{
		RecentStart: make([]*ppzCs.RecentAddressItem, 0),
		RecentDest:  make([]*ppzCs.RecentAddressItem, 0),
		SavedList:   make([]*ppzCs.AddressListItem, 0),
	}

	orders, err := ppzModel.TbPpzOrderModel.ListRecentByUid(uid, 3)
	if err != nil {
		return nil, fmt.Errorf("获取最近订单失败: %w", err)
	}

	// 按经纬度去重
	seenStart := make(map[string]bool)
	seenDest := make(map[string]bool)
	for _, order := range orders {
		if order.StartGaodeData != nil {
			key := fmt.Sprintf("%.6f,%.6f", order.StartGaodeData.Lng, order.StartGaodeData.Lat)
			if !seenStart[key] {
				seenStart[key] = true
				gaodeBytes, _ := json.Marshal(order.StartGaodeData)
				resp.RecentStart = append(resp.RecentStart, &ppzCs.RecentAddressItem{
					GaodeData: gaodeBytes,
				})
			}
		}
		if order.DestGaodeData != nil {
			key := fmt.Sprintf("%.6f,%.6f", order.DestGaodeData.Lng, order.DestGaodeData.Lat)
			if !seenDest[key] {
				seenDest[key] = true
				gaodeBytes, _ := json.Marshal(order.DestGaodeData)
				resp.RecentDest = append(resp.RecentDest, &ppzCs.RecentAddressItem{
					GaodeData: gaodeBytes,
				})
			}
		}
	}

	addresses, err := ppzModel.TbPpzMapAddressModel.ListByUid(uid)
	if err != nil {
		return nil, fmt.Errorf("获取常用地址失败: %w", err)
	}

	for _, addr := range addresses {
		resp.SavedList = append(resp.SavedList, &ppzCs.AddressListItem{
			AddressId: addr.AddressId,
			Remark:    addr.Remark,
			GaodeData: addr.GaodeData,
			Sort:      addr.Sort,
		})
	}

	return resp, nil
}

// 创建订单
func (b *ppzBusiness) CreateOrder(ctx context.Context, uid int64, req *ppzCs.CreateOrderRequest) (*ppzCs.CreateOrderResponse, error) {
	if req.StartGaodeData == "" {
		return nil, errors.New("请选择出发地")
	}
	if req.DestGaodeData == "" {
		return nil, errors.New("请选择目的地")
	}
	if req.DepartTime == "" {
		return nil, errors.New("请选择出发时间")
	}
	if req.PassengerCount <= 0 {
		return nil, errors.New("请选择乘车人数")
	}

	var startGaodeData ppzModel.OrderGaodeData
	if err := json.Unmarshal([]byte(req.StartGaodeData), &startGaodeData); err != nil {
		return nil, fmt.Errorf("出发地数据格式错误: %w", err)
	}

	var destGaodeData ppzModel.OrderGaodeData
	if err := json.Unmarshal([]byte(req.DestGaodeData), &destGaodeData); err != nil {
		return nil, fmt.Errorf("目的地数据格式错误: %w", err)
	}

	activeOrder, _ := ppzModel.TbPpzOrderModel.GetActiveByUid(uid)
	if activeOrder != nil {
		return nil, errors.New("您已有进行中的订单")
	}

	order := &ppzModel.TbPpzOrder{
		Uid:            uid,
		StartGaodeData: &startGaodeData,
		DestGaodeData:  &destGaodeData,
		DepartTime:     req.DepartTime,
		TimeType:       req.TimeType,
		TimeFlex:       req.TimeFlex,
		PassengerCount: req.PassengerCount,
		IsCharter:      req.IsCharter,
	}

	orderId, err := ppzModel.TbPpzOrderModel.Create(order)
	if err != nil {
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	return &ppzCs.CreateOrderResponse{
		OrderId: orderId,
	}, nil
}

// 取消订单
func (b *ppzBusiness) CancelOrder(ctx context.Context, uid int64, req *ppzCs.CancelOrderRequest) (*ppzCs.CancelOrderResponse, error) {
	if req.OrderId <= 0 {
		return nil, errors.New("订单ID无效")
	}

	order, err := ppzModel.TbPpzOrderModel.GetById(req.OrderId)
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	if order.Uid != uid {
		return nil, errors.New("无权操作此订单")
	}

	if order.OrderStatus == ppzModel.OrderStatusCancelled {
		return nil, errors.New("订单已取消")
	}

	if order.OrderStatus >= ppzModel.OrderStatusAccepted {
		return nil, errors.New("司机已接单，无法取消")
	}

	err = ppzModel.TbPpzOrderModel.Cancel(req.OrderId, "乘客主动取消")
	if err != nil {
		return nil, fmt.Errorf("取消订单失败: %w", err)
	}

	return &ppzCs.CancelOrderResponse{}, nil
}

// 获取匹配中的订单列表
func (b *ppzBusiness) GetMatchingOrders(ctx context.Context, uid int64) (*ppzCs.GetMatchingOrdersResponse, error) {
	orders, err := ppzModel.TbPpzOrderModel.ListMatchingByUid(uid)
	if err != nil {
		return nil, fmt.Errorf("获取匹配中订单失败: %w", err)
	}

	list := make([]*ppzCs.OrderBriefItem, 0, len(orders))
	for _, order := range orders {
		list = append(list, &ppzCs.OrderBriefItem{
			OrderId:        order.OrderId,
			StartGaodeData: order.StartGaodeData,
			DestGaodeData:  order.DestGaodeData,
			DepartTime:     order.DepartTime,
			TimeType:       order.TimeType,
			TimeFlex:       order.TimeFlex,
			PassengerCount: order.PassengerCount,
			IsCharter:      order.IsCharter,
			OrderStatus:    order.OrderStatus,
			CreatedAt:      order.CreatedAt,
		})
	}

	return &ppzCs.GetMatchingOrdersResponse{List: list}, nil
}

// 获取我的行程列表
func (b *ppzBusiness) GetMyTrips(ctx context.Context, uid int64) (*ppzCs.GetMyTripsResponse, error) {
	tripIds, err := ppzModel.TbPpzOrderModel.ListActiveTripIdsByUid(uid)
	if err != nil {
		return nil, fmt.Errorf("获取行程列表失败: %w", err)
	}

	list := make([]*ppzCs.TripItem, 0, len(tripIds))
	for _, tripId := range tripIds {
		trip, err := ppzModel.TbPpzTripModel.GetById(tripId)
		if err != nil {
			continue
		}

		orders, err := ppzModel.TbPpzOrderModel.ListByTripIdAndUid(tripId, uid)
		if err != nil {
			continue
		}

		tripOrders := make([]*ppzCs.TripOrderItem, 0, len(orders))
		for _, order := range orders {
			tripOrders = append(tripOrders, &ppzCs.TripOrderItem{
				OrderId:        order.OrderId,
				StartGaodeData: order.StartGaodeData,
				DestGaodeData:  order.DestGaodeData,
				DepartTime:     order.DepartTime,
				TimeType:       order.TimeType,
				TimeFlex:       order.TimeFlex,
				PassengerCount: order.PassengerCount,
				IsCharter:      order.IsCharter,
				OrderStatus:    order.OrderStatus,
			})
		}

		list = append(list, &ppzCs.TripItem{
			TripId: trip.TripId,
			Status: trip.Status,
			Orders: tripOrders,
		})
	}

	return &ppzCs.GetMyTripsResponse{List: list}, nil
}
