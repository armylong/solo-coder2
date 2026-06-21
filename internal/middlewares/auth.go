package middlewares

import (
	"context"
	"time"

	"github.com/armylong/armylong-go/internal/model/user"
	libAuth "github.com/armylong/go-library/service/auth"
	"github.com/armylong/go-library/service/longgin"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT自定义Claims
type Claims struct {
	UID            int64  `json:"uid"`             // 用户ID
	Name           string `json:"name"`            // 用户名
	DeviceType     string `json:"device_type"`     // 设备类型
	UserPermission int    `json:"user_permission"` // 用户权限等级
	jwt.RegisteredClaims
}

// 登录用户信息
type LoginUserInfo struct {
	*user.TbUser
	UserPermission int // 用户权限等级
}

var (
	loginUserInfoCache = libAuth.NewSyncMapCache[*LoginUserInfo]()
	jwtConfig          = libAuth.JWTConfig{
		Secret: "armylong-secret-key-2024",
		Issuer: "armylong",
	}
	middlewareConfig = libAuth.MiddlewareConfig[*LoginUserInfo]{
		JWTConfig:    jwtConfig,
		Cache:        loginUserInfoCache,
		UserLoader:   LoadUserByToken,
		GetUID:       func(u *LoginUserInfo) int64 { return u.Uid },
		IsZero:       func(u *LoginUserInfo) bool { return u == nil },
		Unauthorized: libAuth.DefaultUnauthorized,
	}
)

// 生成JWT Token
func GenerateToken(uid int64, name string, deviceType string, userPermission int, expireDuration time.Duration) (string, int64, error) {
	expireAt := time.Now().Add(expireDuration)
	claims := &Claims{
		UID:            uid,
		Name:           name,
		DeviceType:     deviceType,
		UserPermission: userPermission,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtConfig.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", 0, err
	}
	return tokenString, expireAt.Unix(), nil
}

// 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtConfig.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// 根据Token加载用户信息
func LoadUserByToken(token string) (*LoginUserInfo, error) {
	tokenRecord, err := user.TbUserTokenModel.GetByToken(token)
	if err != nil || tokenRecord == nil {
		return nil, err
	}

	if libAuth.IsTokenExpired(tokenRecord.ExpireAt) {
		return nil, nil
	}

	u, err := user.TbUserModel.GetByUid(tokenRecord.Uid)
	if err != nil || u == nil {
		return nil, err
	}

	permission := user.TbAdminUserModel.GetUserPermission(u.Uid)
	return &LoginUserInfo{
		TbUser:         u,
		UserPermission: permission,
	}, nil
}

// 登录认证中间件
func Middleware(ctx *gin.Context) {
	libAuth.Middleware(middlewareConfig)(ctx)
}

// 管理员权限中间件
func RequireAdmin(ctx *gin.Context) {
	libAuth.RequirePermission(func(u *LoginUserInfo) bool {
		return u.UserPermission >= user.UserPermissionAdmin
	}, libAuth.DefaultForbidden)(ctx)
}

// 超级管理员权限中间件
func RequireSuperAdmin(ctx *gin.Context) {
	libAuth.RequirePermission(func(u *LoginUserInfo) bool {
		return u.UserPermission >= user.UserPermissionSuperAdmin
	}, libAuth.DefaultForbidden)(ctx)
}

// 获取当前登录用户信息
func GetLoginUser(ctx *gin.Context) *LoginUserInfo {
	return libAuth.GetLoginUser[*LoginUserInfo](ctx)
}

// 获取当前登录用户ID
func GetLoginUID(ctx *gin.Context) int64 {
	return libAuth.GetLoginUID(ctx)
}

// 获取当前登录Token
func GetLoginToken(ctx *gin.Context) string {
	return libAuth.GetLoginToken(ctx)
}

// 设置用户信息缓存
func SetCache(token string, info *LoginUserInfo) {
	loginUserInfoCache.Set(token, info)
}

// 删除用户信息缓存
func DeleteCache(token string) {
	loginUserInfoCache.Delete(token)
}

// 批量删除用户信息缓存
func DeleteCacheByTokens(tokens []string) {
	loginUserInfoCache.DeleteByTokens(tokens)
}

// 按UID清除用户信息缓存
func ClearCacheByUID(uid int64) {
	loginUserInfoCache.ClearByUID(uid, func(u *LoginUserInfo) int64 { return u.Uid })
}

// 获取用户信息缓存
func GetCache(token string) *LoginUserInfo {
	return loginUserInfoCache.Get(token)
}

// 获取JWT配置
func GetJWTConfig() libAuth.JWTConfig {
	return jwtConfig
}

// 从context获取登录用户ID
func GetLoginUIDFromContext(ctx context.Context) int64 {
	ginContext, err := longgin.GetGinContext(ctx)
	if err != nil {
		return 0
	}
	return GetLoginUID(ginContext)
}

// 从context获取登录用户信息
func GetLoginUserFromContext(ctx context.Context) *LoginUserInfo {
	ginContext, err := longgin.GetGinContext(ctx)
	if err != nil {
		return nil
	}
	return GetLoginUser(ginContext)
}
