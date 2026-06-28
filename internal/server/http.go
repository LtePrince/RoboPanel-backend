package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"RoboPanel-backend/internal/camera"
	"RoboPanel-backend/internal/config"
	recordsvc "RoboPanel-backend/internal/record/service"
	robotsvc "RoboPanel-backend/internal/robot/service"
	"RoboPanel-backend/internal/ros"
)

type HttpServer struct {
	cfg    *config.Config
	robot  *robotsvc.RobotService
	record *recordsvc.RecordService
	ros    *ros.Client
}

func NewHttpServer(
	cfg *config.Config,
	robot *robotsvc.RobotService,
	record *recordsvc.RecordService,
	rosClient *ros.Client,
	camCfg camera.Config,
	lc fx.Lifecycle,
) *HttpServer {
	s := &HttpServer{cfg: cfg, robot: robot, record: record, ros: rosClient}

	stateHub := newHub()
	camHub := newCameraHub(camCfg)
	router := s.buildRouter(stateHub, camHub)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go broadcastLoop(rosClient, stateHub)
			go camHub.captureLoop()
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Printf("[server] error: %v\n", err)
				}
			}()
			fmt.Printf("[server] listening on :%d\n", cfg.Server.Port)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return s
}

func (s *HttpServer) buildRouter(hub *wsHub, camHub *cameraHub) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/robot/state", handleQuery(s.robot.GetState))
		v1.GET("/ws/state", wsStateHandler(hub))
		v1.GET("/ws/camera", cameraWsHandler(camHub))

		rec := v1.Group("/record")
		{
			rec.POST("/start", handleJSON(s.record.Start))
			rec.POST("/stop", handleQuery(s.record.Stop))
			rec.GET("/status", handleQuery(s.record.Status))
		}

		v1.GET("/demos", handleQuery(s.record.List))
		v1.DELETE("/demos/:name", handleURI(s.record.Delete))
		v1.GET("/demos/:name/files/:file", s.downloadFile)
	}

	return r
}

func (s *HttpServer) downloadFile(c *gin.Context) {
	path, ok := s.record.FileExists(c.Param("name"), c.Param("file"))
	if !ok {
		c.JSON(http.StatusNotFound, Response{Code: "NOT_FOUND", Message: "file not found"})
		return
	}
	c.File(path)
}

var Module = fx.Module("server",
	fx.Provide(NewHttpServer),
)
