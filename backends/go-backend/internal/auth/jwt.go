// AngelaMos | 2026
// jwt.go

package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"

	"github.com/carterperez-dev/templates/go-backend/internal/config"
	"github.com/carterperez-dev/templates/go-backend/internal/core"
	"github.com/carterperez-dev/templates/go-backend/internal/middleware"
)

type JWTManager struct {
	privateKey jwk.Key
	publicKey  jwk.Key
	publicJWKS jwk.Set
	config     config.JWTConfig
}

func NewJWTManager(cfg config.JWTConfig) (*JWTManager, error) {
	privateKeyPEM, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	privateKey, err := jwk.ParseKey(privateKeyPEM, jwk.WithPEM(true))
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	if setErr := privateKey.Set(jwk.AlgorithmKey, jwa.ES256()); setErr != nil {
		return nil, fmt.Errorf("set algorithm: %w", setErr)
	}

	keyID := uuid.New().String()[:8]
	if setErr := privateKey.Set(jwk.KeyIDKey, keyID); setErr != nil {
		return nil, fmt.Errorf("set key id: %w", setErr)
	}

	publicKey, err := privateKey.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("derive public key: %w", err)
	}

	if setErr := publicKey.Set(jwk.KeyUsageKey, "sig"); setErr != nil {
		return nil, fmt.Errorf("set key usage: %w", setErr)
	}

	publicJWKS := jwk.NewSet()
	if addErr := publicJWKS.AddKey(publicKey); addErr != nil {
		return nil, fmt.Errorf("add key to set: %w", addErr)
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		publicJWKS: publicJWKS,
		config:     cfg,
	}, nil
}

func GenerateKeyPair(privateKeyPath, publicKeyPath string) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	jwkPrivate, err := jwk.Import(privateKey)
	if err != nil {
		return fmt.Errorf("import private key: %w", err)
	}

	keyID := uuid.New().String()[:8]
	if setErr := jwkPrivate.Set(jwk.KeyIDKey, keyID); setErr != nil {
		return fmt.Errorf("set key id: %w", setErr)
	}
	if setErr := jwkPrivate.Set(jwk.AlgorithmKey, jwa.ES256()); setErr != nil {
		return fmt.Errorf("set algorithm: %w", setErr)
	}

	privatePEM, err := jwk.Pem(jwkPrivate)
	if err != nil {
		return fmt.Errorf("encode private key: %w", err)
	}

	if writeErr := os.WriteFile(privateKeyPath, privatePEM, 0o600); writeErr != nil {
		return fmt.Errorf("write private key: %w", writeErr)
	}

	jwkPublic, err := jwkPrivate.PublicKey()
	if err != nil {
		return fmt.Errorf("derive public key: %w", err)
	}

	publicPEM, err := jwk.Pem(jwkPublic)
	if err != nil {
		return fmt.Errorf("encode public key: %w", err)
	}

	//nolint:gosec // G306: public key is intentionally world-readable
	if writeErr := os.WriteFile(publicKeyPath, publicPEM, 0o644); writeErr != nil {
		return fmt.Errorf("write public key: %w", writeErr)
	}

	return nil
}

type AccessTokenClaims struct {
	UserID       string `json:"sub"`
	Role         string `json:"role"`
	Tier         string `json:"tier"`
	TokenVersion int    `json:"token_version"`
}

func (m *JWTManager) CreateAccessToken(
	claims AccessTokenClaims,
) (string, error) {
	now := time.Now()

	token, err := jwt.NewBuilder().
		JwtID(uuid.New().String()).
		Issuer(m.config.Issuer).
		Audience([]string{m.config.Audience}).
		Subject(claims.UserID).
		IssuedAt(now).
		Expiration(now.Add(m.config.AccessTokenExpire)).
		NotBefore(now).
		Claim("role", claims.Role).
		Claim("tier", claims.Tier).
		Claim("token_version", claims.TokenVersion).
		Claim("type", "access").
		Build()
	if err != nil {
		return "", fmt.Errorf("build token: %w", err)
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256(), m.privateKey))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return string(signed), nil
}

func (m *JWTManager) VerifyAccessToken(
	ctx context.Context,
	tokenString string,
) (*middleware.AccessTokenClaims, error) {
	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKey(jwa.ES256(), m.publicKey),
		jwt.WithValidate(true),
		jwt.WithIssuer(m.config.Issuer),
		jwt.WithAudience(m.config.Audience),
	)
	if err != nil {
		if isTokenExpiredError(err) {
			return nil, fmt.Errorf("verify token: %w", core.ErrTokenExpired)
		}
		return nil, fmt.Errorf("verify token: %w", core.ErrTokenInvalid)
	}

	var tokenType string
	if err := token.Get("type", &tokenType); err != nil ||
		tokenType != "access" {
		return nil, fmt.Errorf(
			"verify token: invalid token type: %w",
			core.ErrTokenInvalid,
		)
	}

	subject, ok := token.Subject()
	if !ok || subject == "" {
		return nil, fmt.Errorf(
			"verify token: missing subject: %w",
			core.ErrTokenInvalid,
		)
	}

	var roleStr string
	if err := token.Get("role", &roleStr); err != nil {
		return nil, fmt.Errorf(
			"verify token: missing role claim: %w",
			core.ErrTokenInvalid,
		)
	}

	var tierStr string
	if err := token.Get("tier", &tierStr); err != nil {
		return nil, fmt.Errorf(
			"verify token: missing tier claim: %w",
			core.ErrTokenInvalid,
		)
	}

	var versionFloat float64
	if err := token.Get("token_version", &versionFloat); err != nil {
		return nil, fmt.Errorf(
			"verify token: missing token_version claim: %w",
			core.ErrTokenInvalid,
		)
	}

	return &middleware.AccessTokenClaims{
		UserID:       subject,
		Role:         roleStr,
		Tier:         tierStr,
		TokenVersion: int(versionFloat),
	}, nil
}

func isTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "exp") &&
		strings.Contains(errStr, "not satisfied")
}

func (m *JWTManager) GetJWKSHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")

		if err := json.NewEncoder(w).Encode(m.publicJWKS); err != nil {
			http.Error(
				w,
				"Internal Server Error",
				http.StatusInternalServerError,
			)
			return
		}
	}
}

func (m *JWTManager) GetPublicKey() jwk.Key {
	return m.publicKey
}

func (m *JWTManager) GetKeyID() string {
	var kid string
	//nolint:errcheck // key ID always set during NewJWTManager init
	_ = m.privateKey.Get(jwk.KeyIDKey, &kid)
	return kid
}

type RefreshTokenData struct {
	Token     string
	Hash      string
	ExpiresAt time.Time
	FamilyID  string
}

func (m *JWTManager) CreateRefreshToken(
	userID, familyID string,
) (*RefreshTokenData, error) {
	token, err := core.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	hash := core.HashToken(token)
	expiresAt := time.Now().Add(m.config.RefreshTokenExpire)

	if familyID == "" {
		familyID = uuid.New().String()
	}

	return &RefreshTokenData{
		Token:     token,
		Hash:      hash,
		ExpiresAt: expiresAt,
		FamilyID:  familyID,
	}, nil
}

func (m *JWTManager) VerifyRefreshTokenHash(token, storedHash string) bool {
	return core.CompareTokenHash(token, storedHash)
}
