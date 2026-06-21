package ppz

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	ppzCs "github.com/armylong/armylong-go/internal/cs/ppz"
	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
)

type ppzMapBusiness struct{}

var PpzMapBusiness = &ppzMapBusiness{}

// 上传/更新地址
func (b *ppzMapBusiness) UploadAddress(ctx context.Context, uid int64, req *ppzCs.UploadAddressRequest) (*ppzCs.UploadAddressResponse, error) {
	if req.AddressId > 0 {
		address, err := ppzModel.TbPpzMapAddressModel.GetByUidAndId(uid, req.AddressId)
		if err != nil {
			return nil, fmt.Errorf("地址不存在或无权限操作: %w", err)
		}

		address.Remark = req.Remark
		address.GaodeData = req.GaodeData

		if err := ppzModel.TbPpzMapAddressModel.Update(address); err != nil {
			return nil, fmt.Errorf("更新地址失败: %w", err)
		}

		return &ppzCs.UploadAddressResponse{
			AddressId: address.AddressId,
		}, nil
	}

	newAddress := &ppzModel.TbPpzMapAddress{
		Uid:       uid,
		Remark:    req.Remark,
		GaodeData: req.GaodeData,
		Status:    ppzModel.AddressStatusNormal,
	}

	addressId, err := ppzModel.TbPpzMapAddressModel.Create(newAddress)
	if err != nil {
		return nil, fmt.Errorf("创建地址失败: %w", err)
	}

	return &ppzCs.UploadAddressResponse{
		AddressId: addressId,
	}, nil
}

// 获取地址列表
func (b *ppzMapBusiness) GetAddressList(ctx context.Context, uid int64, req *ppzCs.AddressListRequest) (*ppzCs.AddressListResponse, error) {
	response := &ppzCs.AddressListResponse{
		List: make([]*ppzCs.AddressListItem, 0),
	}

	addresses, err := ppzModel.TbPpzMapAddressModel.ListByUid(uid)
	if err != nil {
		return nil, fmt.Errorf("获取地址列表失败: %w", err)
	}

	for _, addr := range addresses {
		response.List = append(response.List, &ppzCs.AddressListItem{
			AddressId: addr.AddressId,
			Remark:    addr.Remark,
			GaodeData: addr.GaodeData,
			Sort:      addr.Sort,
		})
	}

	return response, nil
}

// 更新地址排序
func (b *ppzMapBusiness) UpdateAddressSort(ctx context.Context, uid int64, req *ppzCs.UpdateAddressSortRequest) (*ppzCs.UpdateAddressSortResponse, error) {
	if req.AddressId <= 0 {
		return nil, fmt.Errorf("地址ID不能为空")
	}

	err := ppzModel.TbPpzMapAddressModel.UpdateSort(uid, req.AddressId, req.NewSort)
	if err != nil {
		return nil, fmt.Errorf("更新排序失败: %w", err)
	}

	return &ppzCs.UpdateAddressSortResponse{}, nil
}

// 删除地址
func (b *ppzMapBusiness) DeleteAddress(ctx context.Context, uid int64, req *ppzCs.DeleteAddressRequest) (*ppzCs.DeleteAddressResponse, error) {
	if req.AddressId <= 0 {
		return nil, fmt.Errorf("地址ID不能为空")
	}

	err := ppzModel.TbPpzMapAddressModel.Delete(uid, req.AddressId)
	if err != nil {
		return nil, fmt.Errorf("删除地址失败: %w", err)
	}

	return &ppzCs.DeleteAddressResponse{}, nil
}

// 获取地址详情
func (b *ppzMapBusiness) GetAddressDetail(ctx context.Context, uid int64, req *ppzCs.GetAddressDetailRequest) (*ppzCs.GetAddressDetailResponse, error) {
	if req.AddressId <= 0 {
		return nil, fmt.Errorf("地址ID不能为空")
	}

	address, err := ppzModel.TbPpzMapAddressModel.GetById(req.AddressId)
	if err != nil {
		return nil, fmt.Errorf("获取地址详情失败: %w", err)
	}

	if address == nil || address.Uid != uid || address.Status != ppzModel.AddressStatusNormal {
		return nil, fmt.Errorf("地址不存在或无权限")
	}

	return &ppzCs.GetAddressDetailResponse{
		AddressId: address.AddressId,
		Remark:    address.Remark,
		GaodeData: address.GaodeData,
		Sort:      address.Sort,
	}, nil
}

// 根据经纬度匹配500米内的常用地址备注
type gaodeDataForLocation struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

func (b *ppzMapBusiness) GetAddressRemark(ctx context.Context, uid int64, lng, lat float64) string {
	addresses, err := ppzModel.TbPpzMapAddressModel.ListByUid(uid)
	if err != nil {
		return ""
	}

	var bestAddr *ppzModel.TbPpzMapAddress
	bestDist := 500.0

	for _, addr := range addresses {
		var gd gaodeDataForLocation
		if err := json.Unmarshal(addr.GaodeData, &gd); err != nil {
			continue
		}
		if gd.Lng == 0 && gd.Lat == 0 {
			continue
		}
		d := haversine(lat, lng, gd.Lat, gd.Lng)
		if d < bestDist {
			bestDist = d
			bestAddr = addr
		}
	}

	if bestAddr != nil {
		return bestAddr.Remark
	}
	return ""
}

// 两点间距离（米），Haversine公式
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
