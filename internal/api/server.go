package api

import (
	"net/http"

	"github.com/aleeeeeexxxx/ssht/internal/config"
	"github.com/aleeeeeexxxx/ssht/internal/logger"
	"github.com/aleeeeeexxxx/ssht/internal/tunnel"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config     *config.Config
	configPath string
	manager    *tunnel.Manager
	router     *gin.Engine
}

type TunnelResponse struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	AuthMethod string `json:"auth_method"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
	LocalHost  string `json:"local_host"`
	LocalPort  int    `json:"local_port"`
	State      string `json:"state"`
	Error      string `json:"error,omitempty"`
}

type CreateTunnelRequest struct {
	Name       string `json:"name" binding:"required"`
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port"`
	User       string `json:"user" binding:"required"`
	AuthMethod string `json:"auth_method" binding:"required"`
	KeyPath    string `json:"key_path,omitempty"`
	Password   string `json:"password,omitempty"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port" binding:"required"`
	LocalHost  string `json:"local_host"`
	LocalPort  int    `json:"local_port" binding:"required"`
}

func NewServer(cfg *config.Config, configPath string, manager *tunnel.Manager) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware())

	s := &Server{
		config:     cfg,
		configPath: configPath,
		manager:    manager,
		router:     router,
	}

	s.setupRoutes()
	return s
}

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Log.Infow("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
		)
	}
}

func (s *Server) setupRoutes() {
	s.router.GET("/tunnels", s.listTunnels)
	s.router.POST("/tunnels", s.createTunnel)
	s.router.GET("/tunnels/:name", s.getTunnel)
	s.router.DELETE("/tunnels/:name", s.deleteTunnel)
	s.router.POST("/tunnels/:name/start", s.startTunnel)
	s.router.POST("/tunnels/:name/stop", s.stopTunnel)
}

func (s *Server) Run(addr string) error {
	logger.Log.Infow("starting http server", "addr", addr)
	return s.router.Run(addr)
}

func (s *Server) listTunnels(c *gin.Context) {
	tunnels := make([]TunnelResponse, 0, len(s.config.Tunnels))
	for _, tc := range s.config.Tunnels {
		t := s.manager.Get(tc.Name)
		resp := TunnelResponse{
			Name:       tc.Name,
			Host:       tc.Host,
			Port:       tc.Port,
			User:       tc.User,
			AuthMethod: tc.AuthMethod,
			RemoteHost: tc.RemoteHost,
			RemotePort: tc.RemotePort,
			LocalHost:  tc.LocalHost,
			LocalPort:  tc.LocalPort,
			State:      s.manager.GetState(tc.Name).String(),
		}
		if t != nil {
			resp.Error = t.ErrorMsg()
		}
		tunnels = append(tunnels, resp)
	}
	c.JSON(http.StatusOK, gin.H{"tunnels": tunnels})
}

func (s *Server) createTunnel(c *gin.Context) {
	var req CreateTunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Port == 0 {
		req.Port = 22
	}
	if req.RemoteHost == "" {
		req.RemoteHost = "0.0.0.0"
	}
	if req.LocalHost == "" {
		req.LocalHost = "127.0.0.1"
	}

	tc := tunnel.Config{
		Name:       req.Name,
		Host:       req.Host,
		Port:       req.Port,
		User:       req.User,
		AuthMethod: req.AuthMethod,
		KeyPath:    req.KeyPath,
		Password:   req.Password,
		RemoteHost: req.RemoteHost,
		RemotePort: req.RemotePort,
		LocalHost:  req.LocalHost,
		LocalPort:  req.LocalPort,
	}

	s.config.Add(tc)
	if err := config.Save(s.configPath, s.config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.manager.Add(tc)
	logger.Log.Infow("tunnel created", "tunnel", tc.Name)

	c.JSON(http.StatusCreated, gin.H{
		"message": "tunnel created",
		"tunnel":  req.Name,
	})
}

func (s *Server) getTunnel(c *gin.Context) {
	name := c.Param("name")
	tc := s.config.Get(name)
	if tc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tunnel not found"})
		return
	}

	t := s.manager.Get(name)
	resp := TunnelResponse{
		Name:       tc.Name,
		Host:       tc.Host,
		Port:       tc.Port,
		User:       tc.User,
		AuthMethod: tc.AuthMethod,
		RemoteHost: tc.RemoteHost,
		RemotePort: tc.RemotePort,
		LocalHost:  tc.LocalHost,
		LocalPort:  tc.LocalPort,
		State:      s.manager.GetState(name).String(),
	}
	if t != nil {
		resp.Error = t.ErrorMsg()
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) deleteTunnel(c *gin.Context) {
	name := c.Param("name")
	if s.config.Get(name) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tunnel not found"})
		return
	}

	s.manager.Remove(name)
	s.config.Remove(name)
	if err := config.Save(s.configPath, s.config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Infow("tunnel deleted", "tunnel", name)
	c.JSON(http.StatusOK, gin.H{"message": "tunnel deleted"})
}

func (s *Server) startTunnel(c *gin.Context) {
	name := c.Param("name")
	if s.config.Get(name) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tunnel not found"})
		return
	}

	if err := s.manager.Start(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Infow("tunnel started", "tunnel", name)
	c.JSON(http.StatusOK, gin.H{"message": "tunnel started"})
}

func (s *Server) stopTunnel(c *gin.Context) {
	name := c.Param("name")
	if s.config.Get(name) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tunnel not found"})
		return
	}

	s.manager.Stop(name)
	logger.Log.Infow("tunnel stopped", "tunnel", name)
	c.JSON(http.StatusOK, gin.H{"message": "tunnel stopped"})
}
