package auth

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var DefaultSigningMethod = jwt.SigningMethodHS256

var ErrUnauthorized = errors.New("unauthorized")
var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Type   Type      `json:"type"`

	jwt.StandardClaims
}

type Service struct {
	method jwt.SigningMethod
	secret []byte

	accessTokenTTL time.Duration

	repo Repository
}

func NewService(method jwt.SigningMethod, secret string, accessTokenTTL time.Duration, repo Repository) *Service {
	return &Service{
		method:         method,
		secret:         []byte(secret),
		accessTokenTTL: accessTokenTTL,
		repo:           repo,
	}
}

func (s *Service) IssueRefreshToken(userID uuid.UUID) (token *Token, signedToken string, err error) {
	token = &Token{
		ID:       uuid.New(),
		Type:     TypeRefreshToken,
		IssuedAt: time.Now(),
		UserID:   userID,
	}

	signedToken, err = s.issueToken(token)

	return token, signedToken, err
}

func (s *Service) IssueAccessToken(refreshTokenID uuid.UUID, userID uuid.UUID) (token *Token, signedToken string, err error) {
	expiresAt := time.Now().Add(s.accessTokenTTL)
	token = &Token{
		ID:        uuid.New(),
		Type:      TypeAccessToken,
		ParentID:  &refreshTokenID,
		IssuedAt:  time.Now(),
		ExpiresAt: &expiresAt,
		UserID:    userID,
	}

	signedToken, err = s.issueToken(token)

	return token, signedToken, err
}

func (s *Service) issueToken(token *Token) (signedToken string, err error) {
	err = s.repo.Create(token)
	if err != nil {
		return "", errors.Wrap(err, "failed to save token")
	}

	claims := &Claims{
		UserID: token.UserID,
		Type:   token.Type,
		StandardClaims: jwt.StandardClaims{
			Id:       token.ID.String(),
			IssuedAt: token.IssuedAt.Unix(),
		},
	}
	if token.ExpiresAt != nil {
		claims.StandardClaims.ExpiresAt = token.ExpiresAt.Unix()
	}

	t := jwt.NewWithClaims(s.method, claims)
	signedToken, err = t.SignedString(s.secret)
	if err != nil {
		return "", errors.Wrap(err, "sign failed")
	}

	return signedToken, nil
}

func (s *Service) FetchToken(token string) (*Token, error) {
	claims := new(Claims)
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrUnauthorized
		}

		return nil, ErrInvalidToken
	}
	if !parsed.Valid {
		return nil, ErrUnauthorized
	}

	id, err := uuid.Parse(claims.Id)
	if err != nil {
		return nil, errors.Wrap(err, "uuid parse failed")
	}

	t, err := s.repo.Get(id)
	if errors.Is(err, ErrNotFound) {
		return nil, errors.Wrap(ErrInvalidToken, "token not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to load from storage")
	}

	return t, nil
}
