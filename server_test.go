package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	var sequence = []struct {
		verb, route string
		status      int
		expected    string
	}{
		// Test pool list and creation
		{"GET", "/pools", 200, "[]"},
		{"GET", "/pools/abc", 404, ""},
		{"POST", "/pools?name=abc&template=.sddd", 201, ""},
		{"GET", "/pools/abc", 200, ""},
		{"POST", "/pools?name=qwe", 400, ""},
		{"POST", "/pools?template=.sddd", 400, ""},
		{"POST", "/pools?name=qwe&template=.bad", 400, "Bad Template String"},
		{"POST", "/pools?name=123&template=.rddddd", 201, ""},
		{"GET", "/pools", 200, `["abc","123"]`},

		// test pool status and that minting keeps state
		// both sequential and random noids
		{"POST", "/pools/abc/mint?n=5", 200, `["000","001","002","003","004"]`},
		{"POST", "/pools/abc/mint?n=5", 200, `["005","006","007","008","009"]`},
		{"POST", "/pools/123/mint?n=5", 200, `["00000","00342","00684","01026","01368"]`},
		// advance past
		{"POST", "/pools/123/advancePast?id=12345", 200, ""},
		{"POST", "/pools/123/mint?n=5", 200, `["12687","13029","13371","13713","14055"]`},
		// open and close
		{"PUT", "/pools/123/close", 200, ""},
		{"POST", "/pools/123/mint?n=5", 400, ""},
		{"PUT", "/pools/123/open", 200, ""},
		{"POST", "/pools/123/mint?n=5", 200, `["14397","14739","15081","15423","15765"]`},
	}
	for _, s := range sequence {
		checkRoute(t, s.verb, s.route, s.status, s.expected)
	}
}

func checkRoute(t *testing.T, verb, route string, status int, expected string) {
	req, err := http.NewRequest(verb, testServer.URL+route, nil)
	if err != nil {
		t.Fatal("Problem creating request", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(route, err)
	}
	if resp.StatusCode != status {
		t.Errorf("%s: Expected status %d and received %d",
			route,
			status,
			resp.StatusCode)
	}
	// All sucessful requests should return JSON bodies
	if resp.StatusCode < 300 {
		if resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("%s: Content-Type is not application/json", route)
		}
	}
	if expected != "" {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(route, err)
		}
		if string(body) != expected+"\n" {
			t.Errorf("%s: Expected body %s, got %s",
				route,
				expected,
				body)
		}
	}
	resp.Body.Close()
}

var testServer *httptest.Server

func init() {
	SetupHandlers(nil)
	testServer = httptest.NewServer(nil)
}
