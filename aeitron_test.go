package aeitron

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientInitialization(t *testing.T) {
	// Empty key rejection
	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error on empty API key, got nil")
	}

	// Valid key initialization with custom options
	client, err := NewClient("atk_live_testkey12345",
		WithBaseURL("https://api-test.aeitron.com"),
		WithAPIVersion("2026-11-01"),
		WithTimeout(10*time.Second),
		WithMaxRetries(2),
		WithProjectID("proj_alpha"),
	)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	if client.transport.baseURL != "https://api-test.aeitron.com" {
		t.Errorf("expected baseURL https://api-test.aeitron.com, got %s", client.transport.baseURL)
	}
	if client.transport.projectID != "proj_alpha" {
		t.Errorf("expected projectID proj_alpha, got %s", client.transport.projectID)
	}
}

func TestInferenceChatCompletion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer nke_live_secret" {
			t.Errorf("missing or invalid Authorization header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("missing User-Agent header")
		}

		var req ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.RoutingStrategy != RoutingLatencyOptimal {
			t.Errorf("expected strategy LATENCY_OPTIMAL, got %s", req.RoutingStrategy)
		}

		w.Header().Set("Content-Type", "application/json")
		resp := ChatCompletionResponse{
			ID:    "cmpl_test_99",
			Model: "aeitron-fast-l4",
			Choices: []CompletionChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: "Hello from Aeitron Edge in Dhaka!",
					},
				},
			},
			Telemetry: CompletionTelemetry{
				Provider: "aeitron-bd-dac-1",
				Cost:     0.00012,
				TTFTMs:   14.2,
				Region:   "bd-dac",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client, err := NewClient("nke_live_secret", WithBaseURL(ts.URL))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &ChatCompletionRequest{
		Model: "aeitron-fast-l4",
		Messages: []ChatMessage{
			{Role: "user", Content: "Ping"},
		},
		RoutingStrategy: RoutingLatencyOptimal,
		Jurisdiction:    JurisdictionBD,
	}

	resp, err := client.Inference.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion failed: %v", err)
	}

	if resp.ID != "cmpl_test_99" {
		t.Errorf("expected ID cmpl_test_99, got %s", resp.ID)
	}
	if resp.Telemetry.Region != "bd-dac" {
		t.Errorf("expected region bd-dac, got %s", resp.Telemetry.Region)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message.Content != "Hello from Aeitron Edge in Dhaka!" {
		t.Errorf("unexpected choice payload: %+v", resp.Choices)
	}
}

func TestDomainsLifecycle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/domains":
			if r.Method == "POST" {
				var req OnboardDomainRequest
				json.NewDecoder(r.Body).Decode(&req)
				json.NewEncoder(w).Encode(DomainRecord{
					ID:            "dom_123",
					ProjectID:     req.ProjectID,
					Hostname:      req.Hostname,
					OriginAddress: req.OriginAddress,
					Status:        "PENDING_VERIFICATION",
				})
			} else {
				json.NewEncoder(w).Encode(map[string]any{
					"domains": []DomainRecord{
						{ID: "dom_123", Hostname: "api.example.com", Status: "ACTIVE"},
					},
				})
			}
		case "/v1/domains/dom_123/verify":
			json.NewEncoder(w).Encode(VerifyDomainResult{
				Verified: true,
				Status:   "ACTIVE",
				Message:  "DNS challenge successfully validated at Dhaka edge",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client, err := NewClient("testkey", WithBaseURL(ts.URL))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx := context.Background()

	// 1. Onboard
	dom, err := client.Domains.Onboard(ctx, &OnboardDomainRequest{
		ProjectID:     "proj_live",
		Hostname:      "api.example.com",
		OriginAddress: "10.0.0.1",
		OriginPort:    443,
	})
	if err != nil {
		t.Fatalf("Domains.Onboard failed: %v", err)
	}
	if dom.ID != "dom_123" || dom.Status != "PENDING_VERIFICATION" {
		t.Errorf("unexpected domain record: %+v", dom)
	}

	// 2. Verify
	verify, err := client.Domains.Verify(ctx, dom.ID)
	if err != nil {
		t.Fatalf("Domains.Verify failed: %v", err)
	}
	if !verify.Verified || verify.Status != "ACTIVE" {
		t.Errorf("expected active verified domain, got %+v", verify)
	}

	// 3. List
	list, err := client.Domains.List(ctx)
	if err != nil {
		t.Fatalf("Domains.List failed: %v", err)
	}
	if len(list) != 1 || list[0].Hostname != "api.example.com" {
		t.Errorf("unexpected domain list: %+v", list)
	}
}

func TestAutomaticRetryOnTransientError(t *testing.T) {
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 3 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit exceeded"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TelemetrySummary{
			GlobalRequestsPerSec: 154000.5,
			EdgeCacheHitRatio:    0.942,
		})
	}))
	defer ts.Close()

	client, err := NewClient("testkey",
		WithBaseURL(ts.URL),
		WithMaxRetries(3),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	summary, err := client.Telemetry.GetSummary(context.Background())
	if err != nil {
		t.Fatalf("expected successful retry, got error: %v", err)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
	if summary.EdgeCacheHitRatio != 0.942 {
		t.Errorf("unexpected telemetry: %+v", summary)
	}
}

func TestWebhookSignatureVerification(t *testing.T) {
	verifier := NewWebhookVerifier()
	secret := "whsec_supersecretkey123"
	payload := []byte(`{"event":"domain.verified","domain_id":"dom_99"}`)

	now := time.Now().Unix()
	timestampStr := fmt.Sprintf("%d", now)
	signedPayload := fmt.Sprintf("%s.%s", timestampStr, string(payload))

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	validSig := hex.EncodeToString(mac.Sum(nil))

	validHeader := fmt.Sprintf("t=%s,v1=%s", timestampStr, validSig)

	// 1. Valid signature passes
	err := verifier.VerifySignature(payload, validHeader, secret, 5*time.Minute)
	if err != nil {
		t.Fatalf("expected valid signature to pass, got: %v", err)
	}

	// 2. Tampered payload fails
	tamperedPayload := []byte(`{"event":"domain.verified","domain_id":"dom_HACKED"}`)
	err = verifier.VerifySignature(tamperedPayload, validHeader, secret, 5*time.Minute)
	if err == nil {
		t.Fatal("expected tampered payload to fail verification, got nil")
	}

	// 3. Expired timestamp fails
	oldTimestamp := fmt.Sprintf("%d", time.Now().Add(-1*time.Hour).Unix())
	oldSignedPayload := fmt.Sprintf("%s.%s", oldTimestamp, string(payload))
	macOld := hmac.New(sha256.New, []byte(secret))
	macOld.Write([]byte(oldSignedPayload))
	oldHeader := fmt.Sprintf("t=%s,v1=%s", oldTimestamp, hex.EncodeToString(macOld.Sum(nil)))

	err = verifier.VerifySignature(payload, oldHeader, secret, 5*time.Minute)
	if err == nil {
		t.Fatal("expected expired timestamp to fail verification, got nil")
	}

	// 4. Incorrect secret fails
	err = verifier.VerifySignature(payload, validHeader, "whsec_wrongsecret", 5*time.Minute)
	if err == nil {
		t.Fatal("expected invalid secret to fail, got nil")
	}
}
