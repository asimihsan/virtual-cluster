/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package proxy

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	target      *url.URL
	processName string
	db          *sql.DB
}

func NewProxy(
	target string,
	processName string,
	db *sql.DB,
) (*Proxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		target:      targetURL,
		processName: processName,
		db:          db,
	}, nil
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(statusCode int) {
	sr.statusCode = statusCode
	sr.ResponseWriter.WriteHeader(statusCode)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read the request body into a buffer
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	// Create new io.ReadCloser from the buffer
	rBody1 := io.NopCloser(bytes.NewReader(bodyBytes))

	// Record the HTTP request into the SQLite table
	headers, _ := json.Marshal(r.Header)
	_, err = p.db.Exec(`
		INSERT INTO http_requests (process_name, method, url, headers, body)
	VALUES (?, ?, ?, ?, ?)`,
		p.processName, r.Method, r.URL.String(), string(headers), string(bodyBytes))
	if err != nil {
		log.Printf("Error recording HTTP request: %v", err)
	}

	// Replace the original request body with the second io.ReadCloser
	r.Body = rBody1

	// Create a statusRecorder to capture the status code
	sr := &statusRecorder{ResponseWriter: w}

	// Pass the statusRecorder to the proxy
	proxy := httputil.NewSingleHostReverseProxy(p.target)
	proxy.ServeHTTP(sr, r)

	// Now you can access the status code using sr.statusCode
	log.Debug().Str("process_name", p.processName).Int("status_code", sr.statusCode).Msg("Captured status code")
}
