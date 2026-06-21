package ppz

import (
	"encoding/json"
	"time"

	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
)

// 获取用户信息-请求
type GetPpzUserInfoRequest struct{}

// 获取用户信息-响应
type GetPpzUserInfoResponse struct {
	CarCount     int `json:"car_count"`     // 车辆数
	Status       int `json:"status"`        // 用户状态: 1-正常 2-已封禁
	DriverStatus int `json:"driver_status"` // 司机状态: 1-正常 2-已封禁
}

// 添加车辆-请求
type AddMyCarRequest struct {
	CarModel           string `json:"car_model" form:"car_model"`                       // 车型
	CarLicensePhoto    string `json:"car_license_photo" form:"car_license_photo"`       // 行驶证照片
	DriverLicensePhoto string `json:"driver_license_photo" form:"driver_license_photo"` // 驾驶证照片
	LicensePlate       string `json:"license_plate" form:"license_plate"`               // 车牌号
	CarColor           string `json:"car_color" form:"car_color"`                       // 车辆颜色
	Seats              int    `json:"seats" form:"seats"`                               // 乘客座位数
	CarPhoto           string `json:"car_photo" form:"car_photo"`                       // 车辆照片
	Description        string `json:"description" form:"description"`                   // 车辆简介
}

// 添加车辆-响应
type AddMyCarResponse struct {
	AuditId int64 `json:"audit_id"` // 审核记录ID
	CarId   int64 `json:"car_id"`   // 车辆ID（审核通过后才有值）
}

// 编辑车辆-请求
type EditMyCarRequest struct {
	AuditId            int64  `json:"audit_id" form:"audit_id"`                         // 审核记录ID
	CarId              int64  `json:"car_id" form:"car_id"`                             // 车辆ID（兼容旧版）
	CarModel           string `json:"car_model" form:"car_model"`                       // 车型
	CarLicensePhoto    string `json:"car_license_photo" form:"car_license_photo"`       // 行驶证照片
	DriverLicensePhoto string `json:"driver_license_photo" form:"driver_license_photo"` // 驾驶证照片
	LicensePlate       string `json:"license_plate" form:"license_plate"`               // 车牌号
	CarColor           string `json:"car_color" form:"car_color"`                       // 车辆颜色
	Seats              int    `json:"seats" form:"seats"`                               // 乘客座位数
	CarPhoto           string `json:"car_photo" form:"car_photo"`                       // 车辆照片
	Description        string `json:"description" form:"description"`                   // 车辆简介
}

// 编辑车辆-响应
type EditMyCarResponse struct {
	AuditId int64 `json:"audit_id"` // 审核记录ID
	CarId   int64 `json:"car_id"`   // 车辆ID
}

// 我的车辆列表-请求
type GetMyCarsRequest struct {
	ReviewStatus int `json:"review_status" form:"review_status"` // 审核状态（兼容旧版）
	AuditStatus  int `json:"audit_status" form:"audit_status"`   // 审核状态: 0-全部 1-待审核 2-已通过 3-已驳回
}

// 车辆审核详情
type CarAuditDetail struct {
	AuditId            int64                  `json:"audit_id"`             // 审核记录ID
	CarId              int64                  `json:"car_id"`               // 车辆ID（审核通过后才有值）
	Uid                int64                  `json:"uid"`                  // 用户ID
	CarModel           string                 `json:"car_model"`            // 车型
	CarLicensePhoto    string                 `json:"car_license_photo"`    // 行驶证照片
	DriverLicensePhoto string                 `json:"driver_license_photo"` // 驾驶证照片
	LicensePlate       string                 `json:"license_plate"`        // 车牌号
	CarColor           string                 `json:"car_color"`            // 车辆颜色
	Seats              int                    `json:"seats"`                // 乘客座位数
	CarPhoto           string                 `json:"car_photo"`            // 车辆照片
	Description        string                 `json:"description"`          // 车辆简介
	AuditStatus        int                    `json:"audit_status"`         // 审核状态: 0-删除 1-待审核 2-已通过 3-已驳回
	ReviewStatus       int                    `json:"review_status"`        // 兼容旧版，等于audit_status
	AuditReason        string                 `json:"audit_reason"`         // 审核理由
	AuditData          *ppzModel.CarAuditData `json:"audit_data"`           // 提交的车辆数据
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// 我的车辆列表-响应
type GetMyCarsResponse struct {
	List []*ppzModel.TbPpzCarAudit `json:"list"` // 车辆列表
}

// 删除车辆-请求
type DeleteMyCarRequest struct {
	AuditId int64 `json:"audit_id" form:"audit_id"` // 审核记录ID
	CarId   int64 `json:"car_id" form:"car_id"`     // 车辆ID（兼容旧版）
}

// 删除车辆-响应
type DeleteMyCarResponse struct{}

// 检查是否认证司机-响应
type CheckDriverResponse struct {
	IsCertified bool `json:"is_certified"` // 是否为认证司机
}

// 概览统计-请求
type OverviewStatsRequest struct{}

// 概览统计-响应
type OverviewStatsResponse struct {
	TotalDrivers  int64 `json:"total_drivers"`  // 司机总数
	BannedDrivers int64 `json:"banned_drivers"` // 封禁司机数
	TotalCars     int64 `json:"total_cars"`     // 车辆总数
}

// 司机列表-请求
type DriverListRequest struct {
	Page     int `json:"page" form:"page"`           // 页码，从1开始
	PageSize int `json:"page_size" form:"page_size"` // 每页条数
	Status   int `json:"status" form:"status"`       // 状态: 0-全部 1-正常 2-已封禁
}

// 司机信息
type DriverInfo struct {
	Uid              int64            `json:"uid"`                // 用户ID
	Account          string           `json:"account"`            // 账号
	Name             string           `json:"name"`               // 姓名
	Status           int              `json:"status"`             // 状态: 1-正常 2-已封禁
	BanReason        string           `json:"ban_reason"`         // 封禁原因
	BannedAt         time.Time        `json:"banned_at"`          // 封禁时间
	CarCount         int64            `json:"car_count"`          // 车辆数
	ApprovedCarCount int64            `json:"approved_car_count"` // 审核通过车辆数
	Cars             []*CarSimpleInfo `json:"cars"`               // 车辆列表
}

// 车辆简要信息
type CarSimpleInfo struct {
	CarId        int64  `json:"car_id"`        // 车辆ID
	CarModel     string `json:"car_model"`     // 车型
	AuditStatus  int    `json:"audit_status"`  // 审核状态
	ReviewStatus int    `json:"review_status"` // 兼容旧版
}

// 司机列表-响应
type DriverListResponse struct {
	Drivers  []*DriverInfo `json:"drivers"`   // 司机列表
	Total    int64         `json:"total"`     // 总数
	Page     int           `json:"page"`      // 当前页码
	PageSize int           `json:"page_size"` // 每页条数
}

// 封禁司机-请求
type BanDriverRequest struct {
	Uid       int64  `json:"uid" form:"uid"`               // 用户ID
	BanReason string `json:"ban_reason" form:"ban_reason"` // 封禁原因
}

// 封禁司机-响应
type BanDriverResponse struct{}

// 解封司机-请求
type UnbanDriverRequest struct {
	Uid int64 `json:"uid" form:"uid"` // 用户ID
}

// 解封司机-响应
type UnbanDriverResponse struct{}

// 车辆详情-请求
type GetCarDetailRequest struct {
	AuditId int64 `json:"audit_id" form:"audit_id"` // 审核记录ID
	CarId   int64 `json:"car_id" form:"car_id"`     // 车辆ID（兼容旧版）
}

// 车辆详情-响应
type GetCarDetailResponse struct {
	Car *CarAuditDetail `json:"car"` // 车辆详情
}

// 车辆审核统计-请求
type CarAuditOverviewStatsRequest struct{}

// 车辆审核统计-响应
type CarAuditOverviewStatsResponse struct {
	PendingCount  int64 `json:"pending_count"`  // 待审核数
	ApprovedCount int64 `json:"approved_count"` // 已通过数
	RejectedCount int64 `json:"rejected_count"` // 已驳回数
}

// 车辆审核列表-请求
type CarAuditListRequest struct {
	Page        int    `json:"page" form:"page"`                   // 页码，从1开始
	PageSize    int    `json:"page_size" form:"page_size"`         // 每页条数
	AuditStatus int    `json:"audit_status" form:"audit_status"`   // 审核状态: 0-全部 1-待审核 2-已通过 3-已驳回
	Uid         int64  `json:"uid" form:"uid"`                     // 用户ID筛选
	Account     string `json:"account" form:"account"`             // 账号筛选
	Name        string `json:"name" form:"name"`                   // 姓名筛选
	Phone       string `json:"phone" form:"phone"`                 // 手机号筛选
}

// 审核列表中的司机项
type CarAuditDriverItem struct {
	Uid              int64           `json:"uid"`                // 用户ID
	Account          string          `json:"account"`            // 账号
	Name             string          `json:"name"`               // 姓名
	Phone            string          `json:"phone"`              // 手机号
	PendingCarCount  int             `json:"pending_car_count"`  // 待审核车辆数
	ApprovedCarCount int             `json:"approved_car_count"` // 已通过车辆数
	RejectedCarCount int             `json:"rejected_car_count"` // 已驳回车辆数
	Cars             []*CarAuditItem `json:"cars"`               // 车辆列表
}

// 审核列表中的车辆项
type CarAuditItem struct {
	AuditId     int64     `json:"audit_id"`     // 审核记录ID
	CarId       int64     `json:"car_id"`       // 车辆ID
	CarModel    string    `json:"car_model"`    // 车型
	Seats       int       `json:"seats"`        // 座位数
	AuditStatus int       `json:"audit_status"` // 审核状态
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 车辆审核列表-响应
type CarAuditListResponse struct {
	Drivers  []*CarAuditDriverItem `json:"drivers"`   // 司机列表（按司机聚合）
	Total    int64                 `json:"total"`     // 总数
	Page     int                   `json:"page"`      // 当前页码
	PageSize int                   `json:"page_size"` // 每页条数
}

// 审核通过-请求
type ApproveCarAuditRequest struct {
	AuditId     int64  `json:"audit_id" form:"audit_id"`         // 审核记录ID
	AuditReason string `json:"audit_reason" form:"audit_reason"` // 审核理由
}

// 审核通过-响应
type ApproveCarAuditResponse struct{}

// 审核驳回-请求
type RejectCarAuditRequest struct {
	AuditId     int64  `json:"audit_id" form:"audit_id"`         // 审核记录ID
	AuditReason string `json:"audit_reason" form:"audit_reason"` // 审核理由
}

// 审核驳回-响应
type RejectCarAuditResponse struct{}

// 地址选择器数据-请求
type AddressPickerDataRequest struct{}

// 最近地址项
type RecentAddressItem struct {
	GaodeData json.RawMessage `json:"gaode_data"` // 高德地址数据
}

// 地址选择器数据-响应
type AddressPickerDataResponse struct {
	RecentStart []*RecentAddressItem `json:"recent_start"` // 最近出发地
	RecentDest  []*RecentAddressItem `json:"recent_dest"`  // 最近目的地
	SavedList   []*AddressListItem  `json:"saved_list"`   // 收藏地址
}

// 创建订单-请求
type CreateOrderRequest struct {
	StartGaodeData string `json:"start_gaode_data" form:"start_gaode_data"` // 出发地高德数据
	DestGaodeData  string `json:"dest_gaode_data" form:"dest_gaode_data"`   // 目的地高德数据
	DepartTime     string `json:"depart_time" form:"depart_time"`           // 出发时间
	TimeType       int    `json:"time_type" form:"time_type"`               // 时间类型: 1-出发 2-到达
	TimeFlex       int    `json:"time_flex" form:"time_flex"`               // 时间灵活度(分钟)
	PassengerCount int    `json:"passenger_count" form:"passenger_count"`   // 乘客数
	IsCharter      int    `json:"is_charter" form:"is_charter"`             // 是否包车: 0-否 1-是
}

// 创建订单-响应
type CreateOrderResponse struct {
	OrderId int64 `json:"order_id"` // 订单ID
}

// 取消订单-请求
type CancelOrderRequest struct {
	OrderId int64 `json:"order_id" form:"order_id"` // 订单ID
}

// 取消订单-响应
type CancelOrderResponse struct{}

// 匹配中订单-请求
type GetMatchingOrdersRequest struct{}

// 订单简要信息
type OrderBriefItem struct {
	OrderId        int64                  `json:"order_id"`          // 订单ID
	StartGaodeData *ppzModel.OrderGaodeData `json:"start_gaode_data"` // 出发地
	DestGaodeData  *ppzModel.OrderGaodeData `json:"dest_gaode_data"`  // 目的地
	DepartTime     string                 `json:"depart_time"`       // 出发时间
	TimeType       int                    `json:"time_type"`         // 时间类型
	TimeFlex       int                    `json:"time_flex"`         // 时间弹性
	PassengerCount int                    `json:"passenger_count"`   // 乘客数
	IsCharter      int                    `json:"is_charter"`        // 是否包车
	OrderStatus    int                    `json:"order_status"`      // 订单状态
	CreatedAt      time.Time              `json:"created_at"`
}

// 匹配中订单-响应
type GetMatchingOrdersResponse struct {
	List []*OrderBriefItem `json:"list"` // 匹配中的订单列表
}

// 我的行程-请求
type GetMyTripsRequest struct{}

// 行程中的订单项
type TripOrderItem struct {
	OrderId        int64                  `json:"order_id"`          // 订单ID
	StartGaodeData *ppzModel.OrderGaodeData `json:"start_gaode_data"` // 出发地
	DestGaodeData  *ppzModel.OrderGaodeData `json:"dest_gaode_data"`  // 目的地
	DepartTime     string                 `json:"depart_time"`       // 出发时间
	TimeType       int                    `json:"time_type"`         // 时间类型
	TimeFlex       int                    `json:"time_flex"`         // 时间弹性
	PassengerCount int                    `json:"passenger_count"`   // 乘客数
	IsCharter      int                    `json:"is_charter"`        // 是否包车
	OrderStatus    int                    `json:"order_status"`      // 订单状态
}

// 行程项
type TripItem struct {
	TripId   int64            `json:"trip_id"`   // 行程ID
	Status   int              `json:"status"`    // 行程状态
	Orders   []*TripOrderItem `json:"orders"`    // 行程下的订单
}

// 我的行程-响应
type GetMyTripsResponse struct {
	List []*TripItem `json:"list"` // 行程列表
}
