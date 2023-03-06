// Copyright 2021 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

// Package token contains functions to request a token from the RSC platform.
package token

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type token struct {
	jwtToken *jwt.Token
}

// expired returns true if the token has expired or if the token has no
// expiration time associated with it.
func (t token) expired() bool {
	if t.jwtToken == nil {
		return true
	}

	claims, ok := t.jwtToken.Claims.(jwt.MapClaims)
	if ok {
		// Compare the expiry to 1 minute into the future to avoid the token
		// expiring in transit or because clocks being skewed.
		now := time.Now().Add(1 * time.Minute)
		return !claims.VerifyExpiresAt(now.Unix(), true)
	}

	return true
}

// setAsAuthHeader adds an Authorization header with a bearer token to the
// specified request.
func (t token) setAsAuthHeader(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.jwtToken.Raw))
}

// fromJWT returns a new token from the JWT token text. Note that the token
// signature is not verified.
func fromJWT(text string) (token, error) {
	p := jwt.Parser{}
	jwtToken, _, err := p.ParseUnverified(text, jwt.MapClaims{})
	if err != nil {
		return token{}, fmt.Errorf("failed to parse JWT token: %v", err)
	}

	return token{jwtToken: jwtToken}, nil
}

// Source is used to obtain access tokens from a remote source.
type Source interface {
	token(ctx context.Context) (token, error)
}
