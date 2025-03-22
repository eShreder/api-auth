package auth

import "time"

func (m *JWTManager) GetTokenTTL() time.Duration {
	return m.tokenTTL
} 