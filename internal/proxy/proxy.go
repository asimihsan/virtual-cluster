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
	verbose     bool
}

type ProxyOption func(*Proxy)

func WithVerbose(verbose bool) ProxyOption {
	return func(p *Proxy) {
		p.verbose = verbose
	}
}

func NewProxy(
	target string,
	processName string,
	db *sql.DB,
	opts ...ProxyOption,
) (*Proxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := &Proxy{
		target:      targetURL,
		processName: processName,
		db:          db,
	}
	for _, opt := range opts {
		opt(proxy)
	}
	return proxy, nil
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       bytes.Buffer
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.body.Write(b)
	return rr.ResponseWriter.Write(b)
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
	if p.verbose {
		log.Debug().
			Str("process_name", p.processName).
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Str("body", string(bodyBytes)).
			Msg("Captured HTTP request")
	}

	tx, err := p.db.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec(`
		INSERT INTO http_requests (process_name, method, url, headers, body)
	VALUES (?, ?, ?, ?, ?)`,
		p.processName, r.Method, r.URL.String(), string(headers), string(bodyBytes))
	if err != nil {
		log.Printf("Error recording HTTP request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	requestID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Replace the original request body with the second io.ReadCloser
	r.Body = rBody1

	// Create a responseRecorder to capture the status code, headers and body
	rr := &responseRecorder{ResponseWriter: w, headers: make(http.Header)}

	// Pass the responseRecorder to the proxy
	proxy := httputil.NewSingleHostReverseProxy(p.target)
	proxy.ServeHTTP(rr, r)

	// Record the HTTP response into the SQLite table
	headers, _ = json.Marshal(rr.headers)
	body := rr.body.String()
	if p.verbose {
		log.Debug().
			Str("process_name", p.processName).
			Int("status_code", rr.statusCode).
			Str("headers", string(headers)).
			Str("body", body).
			Msg("Captured HTTP response")
	}

	_, err = p.db.Exec(
		`INSERT INTO http_responses (http_request_id, process_name, status_code, headers, body) VALUES (?, ?, ?, ?, ?)`,
		requestID, p.processName, rr.statusCode, string(headers), body)
	if err != nil {
		log.Printf("Error recording HTTP response: %v", err)
	}
}
