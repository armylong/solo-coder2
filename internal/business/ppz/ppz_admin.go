package ppz

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	ppzCs "github.com/armylong/armylong-go/internal/cs/ppz"
	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
	"github.com/armylong/armylong-go/internal/model/user"
	"github.com/armylong/go-library/service/sqlite"
)

type ppzAdminBusiness struct{}

var PpzAdminBusiness = &ppzAdminBusiness{}

// 概览统计
func (b *ppzAdminBusiness) OverviewStats(ctx context.Context, req *ppzCs.OverviewStatsRequest) (*ppzCs.OverviewStatsResponse, error) {
	totalDrivers, err := ppzModel.TbPpzCarsModel.CountDistinctUids()
	if err != nil {
		totalDrivers = 0
	}

	allUids, err := ppzModel.TbPpzCarsModel.GetAllDistinctUids()
	if err != nil {
		allUids = []int64{}
	}
	bannedDrivers := int64(0)
	for _, uid := range allUids {
		isDriverBanned, _ := ppzModel.TbPpzUserModel.IsDriverBanned(uid)
		if isDriverBanned {
			bannedDrivers++
		}
	}

	totalCars, err := ppzModel.TbPpzCarsModel.CountCars()
	if err != nil {
		totalCars = 0
	}

	return &ppzCs.OverviewStatsResponse{
		TotalDrivers:  totalDrivers,
		BannedDrivers: bannedDrivers,
		TotalCars:     totalCars,
	}, nil
}

// 司机列表
func (b *ppzAdminBusiness) DriverList(ctx context.Context, req *ppzCs.DriverListRequest) (*ppzCs.DriverListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	filteredUids, err := ppzModel.TbPpzCarsModel.GetDistinctUidsByStatus(req.Status)
	if err != nil {
		return nil, fmt.Errorf("获取司机列表失败: %w", err)
	}

	total := int64(len(filteredUids))

	sort.Slice(filteredUids, func(i, j int) bool {
		return filteredUids[i] > filteredUids[j]
	})

	endIdx := int(offset) + req.PageSize
	if endIdx > len(filteredUids) {
		endIdx = len(filteredUids)
	}
	pageUids := filteredUids
	if int(offset) < len(filteredUids) {
		pageUids = filteredUids[int(offset):endIdx]
	} else {
		pageUids = []int64{}
	}

	driverList := make([]*ppzCs.DriverInfo, 0, len(pageUids))

	for _, uid := range pageUids {
		driverRecord, _ := ppzModel.TbPpzUserModel.GetByUid(uid)

		driverStatus := ppzModel.DriverStatusNormal
		banReason := ""
		var bannedAt time.Time

		if driverRecord != nil {
			driverStatus = driverRecord.DriverStatus
			banReason = driverRecord.BanReason
			bannedAt = driverRecord.BannedAt
		}

		userInfo, err := user.TbUserModel.GetByUid(uid)
		if err != nil || userInfo == nil {
			userInfo = &user.TbUser{
				Uid:     uid,
				Account: "未知",
				Name:    "未知",
			}
		}

		carCount, _ := ppzModel.TbPpzCarsModel.CountByUid(uid)

		allCars, _ := ppzModel.TbPpzCarsModel.ListAllByUid(uid)
		approvedCarCount := int64(0)
		carSimpleList := make([]*ppzCs.CarSimpleInfo, 0, len(allCars))

		for _, car := range allCars {
			carSimpleList = append(carSimpleList, &ppzCs.CarSimpleInfo{
				CarId:        car.CarId,
				CarModel:     car.CarModel,
				AuditStatus:  2,
				ReviewStatus: 2,
			})
			approvedCarCount++
		}

		driverList = append(driverList, &ppzCs.DriverInfo{
			Uid:              uid,
			Account:          userInfo.Account,
			Name:             userInfo.Name,
			Status:           driverStatus,
			BanReason:        banReason,
			BannedAt:         bannedAt,
			CarCount:         carCount,
			ApprovedCarCount: approvedCarCount,
			Cars:             carSimpleList,
		})
	}

	return &ppzCs.DriverListResponse{
		Drivers:  driverList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 封禁司机
func (b *ppzAdminBusiness) BanDriver(ctx context.Context, req *ppzCs.BanDriverRequest) (*ppzCs.BanDriverResponse, error) {
	if req.Uid == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	_, _ = ppzModel.TbPpzUserModel.GetOrCreateByUid(req.Uid)

	err := ppzModel.TbPpzUserModel.BanDriver(req.Uid, req.BanReason)
	if err != nil {
		return nil, fmt.Errorf("封禁司机失败: %w", err)
	}

	err = b.rejectAllAuditsByUid(req.Uid)
	if err != nil {
		return nil, fmt.Errorf("封禁成功，但车辆审核状态更新失败: %w", err)
	}

	return nil, nil
}

// 驳回该司机所有待审核/已通过的审核
func (b *ppzAdminBusiness) rejectAllAuditsByUid(uid int64) error {
	var audits []*ppzModel.TbPpzCarAudit
	err := sqlite.DB.Find(ppzModel.TbPpzCarAuditModel.TableName(), &audits, "uid = ? AND audit_status > 0 ORDER BY created_at DESC", uid)
	if err != nil {
		return err
	}

	for _, audit := range audits {
		if audit.AuditStatus == ppzModel.AuditStatusPending || audit.AuditStatus == ppzModel.AuditStatusApproved {
			audit.AuditStatus = ppzModel.AuditStatusRejected
			err := ppzModel.TbPpzCarAuditModel.Update(audit)
			if err != nil {
				continue
			}
			_, _ = ppzModel.TbPpzCarAuditLogModel.CreateFromAudit(audit)
		}
	}

	return nil
}

// 解封司机
func (b *ppzAdminBusiness) UnbanDriver(ctx context.Context, req *ppzCs.UnbanDriverRequest) (*ppzCs.UnbanDriverResponse, error) {
	if req.Uid == 0 {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	driver, err := ppzModel.TbPpzUserModel.GetByUid(req.Uid)
	if err != nil {
		return nil, fmt.Errorf("该司机不存在")
	}

	if driver.DriverStatus != ppzModel.DriverStatusBanned {
		return nil, fmt.Errorf("该司机未被封禁")
	}

	err = ppzModel.TbPpzUserModel.UnbanDriver(req.Uid)
	if err != nil {
		return nil, fmt.Errorf("解封司机失败: %w", err)
	}

	return nil, nil
}

// 车辆详情
func (b *ppzAdminBusiness) GetCarDetail(ctx context.Context, req *ppzCs.GetCarDetailRequest) (*ppzCs.GetCarDetailResponse, error) {
	var audit *ppzModel.TbPpzCarAudit
	var err error

	if req.AuditId > 0 {
		audit, err = ppzModel.TbPpzCarAuditModel.GetById(req.AuditId)
		if err != nil || audit == nil {
			if req.CarId > 0 {
				audit, err = ppzModel.TbPpzCarAuditModel.GetByUidAndCarId(0, req.CarId)
				if err != nil || audit == nil {
					car, err := ppzModel.TbPpzCarsModel.GetById(req.CarId)
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
				return nil, errors.New("车辆不存在或无权限查看")
			}
		}
	} else if req.CarId > 0 {
		audit, err = ppzModel.TbPpzCarAuditModel.GetByUidAndCarId(0, req.CarId)
		if err != nil || audit == nil {
			car, err := ppzModel.TbPpzCarsModel.GetById(req.CarId)
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

// 审核通过（不带原因）
func (b *ppzAdminBusiness) ApproveCarAudit(ctx context.Context, auditId int64) error {
	return PpzAuditBusiness.ApproveCarAudit(ctx, auditId, "")
}

// 审核驳回（不带原因）
func (b *ppzAdminBusiness) RejectCarAudit(ctx context.Context, auditId int64) error {
	return PpzAuditBusiness.RejectCarAudit(ctx, auditId, "")
}

// 车辆审核统计
func (b *ppzAdminBusiness) CarAuditOverviewStats(ctx context.Context, req *ppzCs.CarAuditOverviewStatsRequest) (*ppzCs.CarAuditOverviewStatsResponse, error) {
	pendingCount, _ := ppzModel.TbPpzCarAuditModel.CountByStatus(ppzModel.AuditStatusPending)
	approvedCount, _ := ppzModel.TbPpzCarAuditModel.CountByStatus(ppzModel.AuditStatusApproved)
	rejectedCount, _ := ppzModel.TbPpzCarAuditModel.CountByStatus(ppzModel.AuditStatusRejected)

	return &ppzCs.CarAuditOverviewStatsResponse{
		PendingCount:  pendingCount,
		ApprovedCount: approvedCount,
		RejectedCount: rejectedCount,
	}, nil
}

// 车辆审核列表（按司机聚合）
func (b *ppzAdminBusiness) CarAuditList(ctx context.Context, req *ppzCs.CarAuditListRequest) (*ppzCs.CarAuditListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	allUids, err := ppzModel.TbPpzCarAuditModel.GetDistinctUidsByStatus(req.AuditStatus)
	if err != nil {
		return nil, fmt.Errorf("获取车辆审核列表失败: %w", err)
	}

	filteredUids := make([]int64, 0, len(allUids))
	for _, uid := range allUids {
		if req.Uid > 0 && uid != req.Uid {
			continue
		}

		userInfo, err := user.TbUserModel.GetByUid(uid)
		if err != nil || userInfo == nil {
			continue
		}

		if req.Account != "" && !contains(userInfo.Account, req.Account) {
			continue
		}
		if req.Name != "" && !contains(userInfo.Name, req.Name) {
			continue
		}
		if req.Phone != "" && !contains(userInfo.Phone, req.Phone) {
			continue
		}

		filteredUids = append(filteredUids, uid)
	}

	total := int64(len(filteredUids))

	endIdx := int(offset) + req.PageSize
	if endIdx > len(filteredUids) {
		endIdx = len(filteredUids)
	}
	pageUids := filteredUids
	if int(offset) < len(filteredUids) {
		pageUids = filteredUids[int(offset):endIdx]
	} else {
		pageUids = []int64{}
	}

	driverList := make([]*ppzCs.CarAuditDriverItem, 0, len(pageUids))

	for _, uid := range pageUids {
		userInfo, err := user.TbUserModel.GetByUid(uid)
		if err != nil || userInfo == nil {
			continue
		}

		allAudits, _ := ppzModel.TbPpzCarAuditModel.ListByUidAndStatus(uid, req.AuditStatus)

		pendingCount := 0
		approvedCount := 0
		rejectedCount := 0
		carList := make([]*ppzCs.CarAuditItem, 0, len(allAudits))

		for _, audit := range allAudits {
			carList = append(carList, &ppzCs.CarAuditItem{
				AuditId:     audit.AuditId,
				CarId:       audit.CarId,
				CarModel:    audit.AuditData.CarModel,
				Seats:       audit.AuditData.Seats,
				AuditStatus: audit.AuditStatus,
				CreatedAt:   audit.CreatedAt,
				UpdatedAt:   audit.UpdatedAt,
			})

			switch audit.AuditStatus {
			case ppzModel.AuditStatusPending:
				pendingCount++
			case ppzModel.AuditStatusApproved:
				approvedCount++
			case ppzModel.AuditStatusRejected:
				rejectedCount++
			}
		}

		driverList = append(driverList, &ppzCs.CarAuditDriverItem{
			Uid:              uid,
			Account:          userInfo.Account,
			Name:             userInfo.Name,
			Phone:            userInfo.Phone,
			PendingCarCount:  pendingCount,
			ApprovedCarCount: approvedCount,
			RejectedCarCount: rejectedCount,
			Cars:             carList,
		})
	}

	return &ppzCs.CarAuditListResponse{
		Drivers:  driverList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 审核通过（带原因）
func (b *ppzAdminBusiness) ApproveCarAuditWithReason(ctx context.Context, req *ppzCs.ApproveCarAuditRequest) (*ppzCs.ApproveCarAuditResponse, error) {
	if req.AuditId == 0 {
		return nil, errors.New("审核记录ID不能为空")
	}

	err := PpzAuditBusiness.ApproveCarAudit(ctx, req.AuditId, req.AuditReason)
	if err != nil {
		return nil, fmt.Errorf("审核通过失败: %w", err)
	}

	return &ppzCs.ApproveCarAuditResponse{}, nil
}

// 审核驳回（带原因）
func (b *ppzAdminBusiness) RejectCarAuditWithReason(ctx context.Context, req *ppzCs.RejectCarAuditRequest) (*ppzCs.RejectCarAuditResponse, error) {
	if req.AuditId == 0 {
		return nil, errors.New("审核记录ID不能为空")
	}

	err := PpzAuditBusiness.RejectCarAudit(ctx, req.AuditId, req.AuditReason)
	if err != nil {
		return nil, fmt.Errorf("审核驳回失败: %w", err)
	}

	return &ppzCs.RejectCarAuditResponse{}, nil
}

// 字符串包含
func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0)
}

// 子串查找
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
