package mw

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/jwt"
	db "offer_tiktok/biz/dal/db"
	"offer_tiktok/biz/model/basic/user"
	_ "offer_tiktok/pkg/errno"
	"offer_tiktok/pkg/utils"
	"strconv"
	"time"
)

var JwtMiddleware *jwt.HertzJWTMiddleware
var identity = "user_id"

func Init() {
	JwtMiddleware, _ = jwt.New(&jwt.HertzJWTMiddleware{
		Key:         []byte("tiktok secret key"),
		TokenLookup: "query:token,form:token",
		Timeout:     24 * time.Hour,
		IdentityKey: identity,
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var loginRequest user.DouyinUserLoginRequest
			if err := c.BindAndValidate(&loginRequest); err != nil {
				return nil, err
			}
			password, err := utils.MD5(loginRequest.Password)
			user_id, err := db.VerifyUser(loginRequest.Username, password)
			if err != nil {
				return nil, err
			}
			c.Set("user_id", user_id)
			return user_id, nil
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(int64); ok {
				return jwt.MapClaims{
					jwt.IdentityKey: v,
				}
			}
			return jwt.MapClaims{}
		},
		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
			hlog.CtxInfof(ctx, "Login success ，token is issued clientIP: "+c.ClientIP())
			c.Set("token", token)
		},

		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			if v, ok := data.(float64); ok {
				v := int64(v)
				raw_id := c.Query("user_id")
				if len(raw_id) == 0 {
					if ok, _ := db.QueryUserID(v); ok {
						c.Set("user_id", v)
						hlog.CtxInfof(ctx, "Token is verified clientIP: "+c.ClientIP())
						return true
					}
					return false
				}
				user_id, _ := strconv.ParseInt(raw_id, 10, 64)
				if v == user_id {
					hlog.CtxInfof(ctx, "Token is verified clientIP: "+c.ClientIP())
					return true
				}
				return false
			}
			return false
		},
	})

}