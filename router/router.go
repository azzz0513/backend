package router

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
	"net/http"
	"time"
	"web_app/controller"
	_ "web_app/docs"
	"web_app/logger"
	"web_app/middlewares"
)

// SetUp 路由管理
// @securityDefinitions.apikey  BearerAuth
// @in header
// @name Authorization
// @description 格式: Bearer <access_token>
func SetUp(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // 设置为发布模式
	}
	r := gin.New()
	// 使用中间件
	r.Use(logger.GinLogger(), logger.GinRecovery(true), middlewares.RateLimitMiddleware(2*time.Second, 1000))

	r.LoadHTMLFiles("templates/index.html")
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	//r.GET("/ping", func(c *gin.Context) {
	//	c.String(http.StatusOK, "pong")
	//})
	v1 := r.Group("/api/v1")
	// 注册业务路由
	v1.POST("/signup", controller.SignUpHandler)

	// 登录业务路由
	v1.POST("/login", controller.LoginHandler)

	// 应用JWT认证中间件，需要登录后使用的业务放到这个中间件后面
	v1.Use(middlewares.JWTAuthMiddleware())

	{
		// 修改当前用户数据
		v1.POST("/change_data", controller.UpdateUserHandler)

		// 创建成员名单
		v1.POST("/create_member_list", controller.CreateMemberListHandler)
		// 往成员名单添加成员
		v1.POST("/add_member", controller.AddMemberHandler)
		// 从成员名单中删除成员
		v1.POST("/delete_member", controller.DeleteMemberHandler)
		// 查看当前用户创建的列表
		v1.GET("/member_list", controller.GetListListHandler)
		// 查看成员名单详情
		v1.GET("/member_list/:id", controller.GetListDetailHandler)

		// 发布打卡活动
		v1.POST("/checkin", controller.CreateCheckinHandler)
		// 参与打卡活动
		v1.POST("/participate/:id", controller.ParticipateHandler)
		// 查看统计数据（长期考勤）（可分别查看前一日，前一周，以及前一个月的数据）

		// 查看创建的打卡活动详情（30天内）
		v1.GET("/checkin/:id", controller.GetCheckinDetailHandler)
		// 查看当前用户需要参加的打卡活动列表（活动未结束）
		v1.GET("/checkin_list", controller.GetCheckinListHandler)
		// 查看当前用户创建的打卡活动列表
		v1.GET("/created_list", controller.GetCreatedCheckinListHandler)
		// 查看已参加过的打卡活动历史记录列表（30天内）
		v1.GET("/history", controller.GetHistoryListHandler)
	}

	// 注册pprof相关路由
	pprof.Register(r)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
