# Aeitron Official Go SDK (`github.com/aeitron/api-go-sdk`)

[![Go Reference](https://pkg.go.dev/badge/github.com/aeitron/api-go-sdk.svg)](https://pkg.go.dev/github.com/aeitron/api-go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Official, zero-external-dependency Go client library for the **Aeitron Intelligent Global Edge & Sovereign AI Platform**.

---

## Features

- **Zero External Runtime Dependencies:** Built strictly using the Go standard library (`net/http`, `crypto/hmac`, `context`, etc.).
- **Multi-Objective AI Orchestration:** Query LLM completions with explicit routing strategies (`LATENCY_OPTIMAL`, `COST_MINIMAL`, `CARBON_NEUTRAL`, `SOVEREIGN_PINNED`).
- **Domain & Edge Lifecycle:** Programmatic onboarding, DNS verification, and origin pool health management.
- **Enterprise Webhook Verification:** Constant-time HMAC-SHA256 signature verification with anti-replay timestamp tolerance.
- **Resilient Reliability:** Automatic exponential backoff with jitter on HTTP 429 and 5xx transient failures.
- **Strictly Typed:** Full compile-time type safety across all platform models.

---

## Installation

```bash
go get github.com/aeitron/api-go-sdk
```

---

## Quickstart

### 1. Initialize Client

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aeitron/api-go-sdk"
)

func main() {
	client, err := aeitron.NewClient("atk_live_your_api_key_here",
		aeitron.WithTimeout(15*time.Second),
		aeitron.WithMaxRetries(3),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Aeitron client: %v", err)
	}

	ctx := context.Background()

	// 2. Chat Completion with Intelligent Multi-Objective Routing
	resp, err := client.Inference.CreateChatCompletion(ctx, &aeitron.ChatCompletionRequest{
		Model: "aeitron-fast-l4",
		Messages: []aeitron.ChatMessage{
			{Role: "user", Content: "Analyze edge latency for Dhaka PoP."},
		},
		RoutingStrategy: aeitron.RoutingLatencyOptimal,
		Jurisdiction:    aeitron.JurisdictionBD,
	})
	if err != nil {
		log.Fatalf("Inference failed: %v", err)
	}

	fmt.Printf("Model Output: %s\n", resp.Choices[0].Message.Content)
	fmt.Printf("Edge Telemetry: Region=%s, TTFT=%.2fms, Cost=$%.5f\n",
		resp.Telemetry.Region, resp.Telemetry.TTFTMs, resp.Telemetry.Cost)
}
```

---

### 2. Verify Webhook Signatures

Prevent forged webhooks and replay attacks with constant-time verification:

```go
verifier := aeitron.NewWebhookVerifier()

err := verifier.VerifySignature(
	rawPayloadBytes,
	r.Header.Get("Aeitron-Signature"),
	"whsec_your_webhook_signing_secret",
	5*time.Minute, // Replay tolerance window
)
if err != nil {
	http.Error(w, "Unauthorized signature", http.StatusUnauthorized)
	return
}
```

---

## Testing

```bash
go test -v -race ./...
```

