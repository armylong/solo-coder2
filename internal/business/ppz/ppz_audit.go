package ppz

import (
	"context"
	"errors"
	"fmt"

	ppzCs "github.com/armylong/armylong-go/internal/cs/ppz"
	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
)

type ppzAuditBusiness struct{}

var PpzAuditBusiness = &ppzAuditBusiness{}

// 检查司机是否被封禁
func (b *ppzAuditBusiness) checkDriverStatus(ctx context.Context, uid int64) error {
	isDriverBanned, err := ppzModel.TbPpzUserModel.IsDriverBanned(uid)
	if err != nil {
		return fmt.Errorf("检查用户状态失败: %w", err)
	}
	if isDriverBanned {
		return errors.New("您的司机状态异常, 不能管理车辆, 如有需要请联系客服")
	}
	return nil
}

// 写一条审核日志
func (b *ppzAuditBusiness) createAuditLog(audit *ppzModel.TbPpzCarAudit) error {
	_, err := ppzModel.TbPpzCarAuditLogModel.CreateFromAudit(audit)
	return err
}

// 添加车辆
func (b *ppzAuditBusiness) AddMyCar(ctx context.Context, uid int64, req *ppzCs.AddMyCarRequest) (*ppzCs.AddMyCarResponse, error) {
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	if err := b.checkDriverStatus(ctx, uid); err != nil {
		return nil, err
	}

	if req.CarModel == "" {
		return nil, errors.New("车辆型号不能为空")
	}
	if req.CarLicensePhoto == "" {
		return nil, errors.New("行驶证照片不能为空")
	}
	if req.DriverLicensePhoto == "" {
		return nil, errors.New("驾驶证照片不能为空")
	}
	if req.LicensePlate == "" {
		return nil, errors.New("车牌号不能为空")
	}
	if req.CarColor == "" {
		return nil, errors.New("车辆颜色不能为空")
	}
	if req.Seats <= 0 {
		return nil, errors.New("乘客座位数必须大于0")
	}
	if req.CarPhoto == "" {
		return nil, errors.New("车辆照片不能为空")
	}
	if req.Description == "" {
		return nil, errors.New("车辆简介不能为空")
	}

	audit := &ppzModel.TbPpzCarAudit{
		Uid:   uid,
		CarId: 0,
		AuditData: &ppzModel.CarAuditData{
			CarModel:           req.CarModel,
			CarLicensePhoto:    req.CarLicensePhoto,
			DriverLicensePhoto: req.DriverLicensePhoto,
			LicensePlate:       req.LicensePlate,
			CarColor:           req.CarColor,
			Seats:              req.Seats,
			CarPhoto:           req.CarPhoto,
			Description:        req.Description,
		},
		AuditStatus: ppzModel.AuditStatusPending,
	}

	auditId, err := ppzModel.TbPpzCarAuditModel.Create(audit)
	if err != nil {
		return nil, fmt.Errorf("添加车辆审核记录失败: %w", err)
	}

	audit.AuditId = auditId
	_ = b.createAuditLog(audit)

	_, _ = ppzModel.TbPpzUserModel.GetOrCreateByUid(uid)

	return &ppzCs.AddMyCarResponse{
		AuditId: auditId,
		CarId:   0,
	}, nil
}

// 获取我的车辆列表
func (b *ppzAuditBusiness) GetMyCars(ctx context.Context, uid int64, req *ppzCs.GetMyCarsRequest) (*ppzCs.GetMyCarsResponse, error) {
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	if err := b.checkDriverStatus(ctx, uid); err != nil {
		return nil, err
	}

	auditStatus := req.AuditStatus
	if auditStatus == 0 && req.ReviewStatus != 0 {
		auditStatus = req.ReviewStatus
	}

	if auditStatus < 0 || auditStatus > 3 {
		auditStatus = 0
	}

	audits, err := ppzModel.TbPpzCarAuditModel.ListByUid(uid, auditStatus)
	if err != nil {
		return nil, fmt.Errorf("获取车辆列表失败: %w", err)
	}

	return &ppzCs.GetMyCarsResponse{
		List: audits,
	}, nil
}

// 编辑车辆
func (b *ppzAuditBusiness) EditMyCar(ctx context.Context, uid int64, req *ppzCs.EditMyCarRequest) (*ppzCs.EditMyCarResponse, error) {
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	if err := b.checkDriverStatus(ctx, uid); err != nil {
		return nil, err
	}

	auditId := req.AuditId
	if auditId == 0 && req.CarId > 0 {
		audit, err := ppzModel.TbPpzCarAuditModel.GetByUidAndCarId(uid, req.CarId)
		if err != nil || audit == nil {
			return nil, errors.New("车辆不存在或无权限编辑")
		}
		auditId = audit.AuditId
	}

	if auditId <= 0 {
		return nil, errors.New("请提供审核ID或车辆ID")
	}

	if req.CarModel == "" {
		return nil, errors.New("车辆型号不能为空")
	}
	if req.CarLicensePhoto == "" {
		return nil, errors.New("行驶证照片不能为空")
	}
	if req.DriverLicensePhoto == "" {
		return nil, errors.New("驾驶证照片不能为空")
	}
	if req.LicensePlate == "" {
		return nil, errors.New("车牌号不能为空")
	}
	if req.CarColor == "" {
		return nil, errors.New("车辆颜色不能为空")
	}
	if req.Seats <= 0 {
		return nil, errors.New("乘客座位数必须大于0")
	}
	if req.CarPhoto == "" {
		return nil, errors.New("车辆照片不能为空")
	}
	if req.Description == "" {
		return nil, errors.New("车辆简介不能为空")
	}

	audit, err := ppzModel.TbPpzCarAuditModel.GetByUidAndId(uid, auditId)
	if err != nil || audit == nil {
		return nil, errors.New("车辆不存在或无权限编辑")
	}

	audit.AuditData = &ppzModel.CarAuditData{
		CarModel:           req.CarModel,
		CarLicensePhoto:    req.CarLicensePhoto,
		DriverLicensePhoto: req.DriverLicensePhoto,
		LicensePlate:       req.LicensePlate,
		CarColor:           req.CarColor,
		Seats:              req.Seats,
		CarPhoto:           req.CarPhoto,
		Description:        req.Description,
	}
	audit.AuditStatus = ppzModel.AuditStatusPending

	err = ppzModel.TbPpzCarAuditModel.Update(audit)
	if err != nil {
		return nil, fmt.Errorf("更新车辆审核记录失败: %w", err)
	}

	_ = b.createAuditLog(audit)

	return &ppzCs.EditMyCarResponse{
		AuditId: audit.AuditId,
		CarId:   audit.CarId,
	}, nil
}

// 删除车辆
func (b *ppzAuditBusiness) DeleteMyCar(ctx context.Context, uid int64, req *ppzCs.DeleteMyCarRequest) (*ppzCs.DeleteMyCarResponse, error) {
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	if err := b.checkDriverStatus(ctx, uid); err != nil {
		return nil, err
	}

	auditId := req.AuditId
	if auditId == 0 && req.CarId > 0 {
		audit, err := ppzModel.TbPpzCarAuditModel.GetByUidAndCarId(uid, req.CarId)
		if err != nil || audit == nil {
			return nil, errors.New("车辆不存在或无权限删除")
		}
		auditId = audit.AuditId
	}

	if auditId <= 0 {
		return nil, errors.New("请提供审核ID或车辆ID")
	}

	audit, err := ppzModel.TbPpzCarAuditModel.GetByUidAndId(uid, auditId)
	if err != nil || audit == nil {
		return nil, errors.New("车辆不存在或无权限删除")
	}

	audit.AuditStatus = ppzModel.AuditStatusDeleted
	err = ppzModel.TbPpzCarAuditModel.Update(audit)
	if err != nil {
		return nil, fmt.Errorf("删除车辆审核记录失败: %w", err)
	}

	_ = b.createAuditLog(audit)

	return nil, nil
}

// 审核通过
func (b *ppzAuditBusiness) ApproveCarAudit(ctx context.Context, auditId int64, auditReason string) error {
	audit, err := ppzModel.TbPpzCarAuditModel.GetById(auditId)
	if err != nil || audit == nil {
		return errors.New("审核记录不存在")
	}

	var carId int64
	if audit.CarId == 0 {
		car := &ppzModel.TbPpzCars{
			Uid:                audit.Uid,
			CarModel:           audit.AuditData.CarModel,
			CarLicensePhoto:    audit.AuditData.CarLicensePhoto,
			DriverLicensePhoto: audit.AuditData.DriverLicensePhoto,
			LicensePlate:       audit.AuditData.LicensePlate,
			CarColor:           audit.AuditData.CarColor,
			Seats:              audit.AuditData.Seats,
			CarPhoto:           audit.AuditData.CarPhoto,
			Description:        audit.AuditData.Description,
			Status:             ppzModel.CarStatusNormal,
		}
		carId, err = ppzModel.TbPpzCarsModel.Create(car)
		if err != nil {
			return fmt.Errorf("创建车辆记录失败: %w", err)
		}
	} else {
		car, err := ppzModel.TbPpzCarsModel.GetById(audit.CarId)
		if err != nil {
			return fmt.Errorf("获取车辆记录失败: %w", err)
		}
		car.CarModel = audit.AuditData.CarModel
		car.CarLicensePhoto = audit.AuditData.CarLicensePhoto
		car.DriverLicensePhoto = audit.AuditData.DriverLicensePhoto
		car.LicensePlate = audit.AuditData.LicensePlate
		car.CarColor = audit.AuditData.CarColor
		car.Seats = audit.AuditData.Seats
		car.CarPhoto = audit.AuditData.CarPhoto
		car.Description = audit.AuditData.Description
		car.Status = ppzModel.CarStatusNormal
		err = ppzModel.TbPpzCarsModel.Update(car)
		if err != nil {
			return fmt.Errorf("更新车辆记录失败: %w", err)
		}
		carId = audit.CarId
	}

	audit.CarId = carId
	audit.AuditStatus = ppzModel.AuditStatusApproved
	audit.AuditReason = auditReason
	err = ppzModel.TbPpzCarAuditModel.Update(audit)
	if err != nil {
		return fmt.Errorf("更新审核记录状态失败: %w", err)
	}

	_ = b.createAuditLog(audit)

	_ = ppzModel.TbPpzUserModel.IncrCarCount(audit.Uid, 1)

	return nil
}

// 审核驳回
func (b *ppzAuditBusiness) RejectCarAudit(ctx context.Context, auditId int64, auditReason string) error {
	audit, err := ppzModel.TbPpzCarAuditModel.GetById(auditId)
	if err != nil || audit == nil {
		return errors.New("审核记录不存在")
	}

	audit.AuditStatus = ppzModel.AuditStatusRejected
	audit.AuditReason = auditReason
	err = ppzModel.TbPpzCarAuditModel.Update(audit)
	if err != nil {
		return fmt.Errorf("更新审核记录状态失败: %w", err)
	}

	_ = b.createAuditLog(audit)

	return nil
}

// 检查是否认证司机
func (b *ppzAuditBusiness) CheckDriver(ctx context.Context, uid int64) (*ppzCs.CheckDriverResponse, error) {
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	isCertified, err := ppzModel.TbPpzCarAuditModel.HasApprovedAudit(uid)
	if err != nil {
		return nil, fmt.Errorf("检查失败: %w", err)
	}

	isDriverBanned, _ := ppzModel.TbPpzUserModel.IsDriverBanned(uid)
	if isDriverBanned {
		isCertified = false
	}

	return &ppzCs.CheckDriverResponse{
		IsCertified: isCertified,
	}, nil
}
