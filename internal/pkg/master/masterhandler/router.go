package masterhandler

import "github.com/gin-gonic/gin"

func RegisterRouters(r *gin.Engine) {

	configRoute(r)
}

func configRoute(r *gin.Engine) {

	hello := r.Group("/ping")
	{
		hello.GET("", func(c *gin.Context) {
			c.JSON(200, "pong")
		})
	}

	job := r.Group("/job")
	{
		job.POST("add", defaultJobRouter.CreateOrUpdate)
		job.POST("del", defaultJobRouter.Delete)
		job.GET("find", defaultJobRouter.FindById)
		job.POST("search", defaultJobRouter.Search)
		job.POST("log", defaultJobRouter.SearchLog)
	}

	stat := r.Group("/statis")
	{
		stat.GET("today", defaultStatRouter.GetTodayStatistics)
		stat.GET("week", defaultStatRouter.GetWeekStatistics)
		stat.GET("system", defaultStatRouter.GetSystemInfo)

	}

	node := r.Group("/node")
	{
		node.POST("search", defaultNodeRouter.Search)
		node.POST("del", defaultNodeRouter.Delete)
	}

}
