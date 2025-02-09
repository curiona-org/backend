package api

func (s *Server) setupRoutes() {
	s.instance.GET("/health", s.handler.HealthCheck)

	s.instance.POST("/auth", s.handler.Auth)
	s.instance.GET("/profile", s.handler.GetProfile, s.handler.MiddlewareAuth)

	s.instance.GET("/roadmaps", s.handler.ListUserRoadmaps, s.handler.MiddlewareAuth)
	s.instance.GET("/roadmaps/:slug", s.handler.GetRoadmapBySlug, s.handler.MiddlewareAuth)
	s.instance.POST("/roadmaps", s.handler.GenerateRoadmap, s.handler.MiddlewareAuth)
	s.instance.GET("/roadmaps/topic/:slug", s.handler.GetTopicBySlug, s.handler.MiddlewareAuth)
}
