// Copyright 2025 Rubrik, Inc.
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

package testnet

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
)

var dummyCrtPem = []byte(`
-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----
`)

var dummyKeyPem = []byte(`
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----
`)

// ServeWithTLS serves the handler function over HTTPS by accepting incoming
// requests.  Intended to be used with a pipenet in unit tests.
func ServeWithTLS(lis net.Listener, handler HandlerFunc) CancelFunc {
	return serve(lis, serveHTTPS, handler)
}

// ServeJSONWithTLS serves the handler function over HTTPS by accepting incoming
// request. The response content-type is set to application/json.  Intended to
// be used with a pipenet in unit tests.
func ServeJSONWithTLS(lis net.Listener, handler HandlerFunc) CancelFunc {
	return ServeWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		return handler(w, req)
	})
}

// serveHTTPS starts an HTTP server using TLS with a dummy certificate.
// Note, if the certificate or its private key is invalid the function will
// panic.
func serveHTTPS(lis net.Listener, server *http.Server) error {
	cert, err := tls.X509KeyPair(dummyCrtPem, dummyKeyPem)
	if err != nil {
		panic(fmt.Sprintf("failed to create dummy key pair: %s", err))
	}

	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	if err := server.ServeTLS(lis, "", ""); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return err
}
