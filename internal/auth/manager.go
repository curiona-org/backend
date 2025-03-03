package auth

type Manager struct {
	cfg *ManagerConfig
}

func NewManager(cfg *ManagerConfig) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

func (m *Manager) NewAccessToken(id int) *Token {
	return NewToken(m.cfg.AccessSecretKey, id, m.cfg.AccessExpiresIn)
}

func (m *Manager) NewRefreshToken(id int) *Token {
	return NewToken(m.cfg.RefreshSecretKey, id, m.cfg.RefreshExpiresIn)
}

func (m *Manager) VerifyAccessToken(tokenStr string) (*Token, error) {
	t := NewToken(m.cfg.AccessSecretKey, 0, m.cfg.AccessExpiresIn)
	return t.Unmarshal(tokenStr)
}

func (m *Manager) VerifyRefreshToken(tokenStr string) (*Token, error) {
	t := NewToken(m.cfg.RefreshSecretKey, 0, m.cfg.RefreshExpiresIn)
	return t.Unmarshal(tokenStr)
}
