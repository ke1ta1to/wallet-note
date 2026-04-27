package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ke1ta1to/wallet-note/internal/platform/apperror"
)

type Claims struct {
	Sub      string
	Username string
}

func ExtractClaims(r *http.Request) (*Claims, error) {
	raw := r.Header.Get("x-amzn-request-context")
	if raw == "" {
		return nil, apperror.ErrUnauthorized
	}

	var ctx requestContext
	if err := json.Unmarshal([]byte(raw), &ctx); err != nil {
		return nil, fmt.Errorf("parse x-amzn-request-context: %w", err)
	}

	sub, _ := ctx.Authorizer.Jwt.Claims["sub"].(string)
	if sub == "" {
		return nil, apperror.ErrUnauthorized
	}
	username, _ := ctx.Authorizer.Jwt.Claims["username"].(string)

	return &Claims{Sub: sub, Username: username}, nil
}

// requestContext mirrors the subset we read from API Gateway HTTP API payload v2.0.
// https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html
type requestContext struct {
	Authorizer struct {
		Jwt struct {
			Claims map[string]any `json:"claims"`
		} `json:"jwt"`
	} `json:"authorizer"`
}
