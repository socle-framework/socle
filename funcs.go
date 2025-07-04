package socle

import "github.com/socle-framework/socle/pkg/auth"

func (s *Socle) GenerateApiToken(username string) (string, *auth.Payload, error) {
	tokenString, payload, err := s.Authenticator.GenerateToken(
		username,
		s.env.auth.token.exp,
		s.env.auth.token.iss,
	)

	if err != nil {
		return "", nil, err
	}
	return tokenString, payload, nil
}

func (s *Socle) GenerateRefreshToken(username string) (string, *auth.Payload, error) {
	tokenString, payload, err := s.Authenticator.GenerateToken(
		username,
		s.env.auth.token.refresh,
		s.env.auth.token.iss,
	)

	if err != nil {
		return "", nil, err
	}
	return tokenString, payload, nil
}
