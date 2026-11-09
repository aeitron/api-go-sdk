package aeitron

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Client is the primary entrypoint for the Aeitron Go SDK.
type Client struct {
	transport *transport

	Inference *InferenceService
	Domains   *DomainsService
	DNS       *DNSService
	Telemetry *TelemetryService
	Compute   *ComputeService
	Storage   *StorageService
	Network   *NetworkService
	Webhooks  *WebhookVerifier
}

// NewClient constructs an authenticated Aeitron API client.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	trimmedKey := strings.TrimSpace(apiKey)
	if trimmedKey == "" {
		return nil, errors.New("aeitron: API key cannot be empty (must provide nke_live_... or atk_live_...)")
	}

	cfg := defaultOptions()
	for _, opt := range opts {
		opt(cfg)
	}

	tp := newTransport(trimmedKey, cfg)

	c := &Client{
		transport: tp,
		Webhooks:  NewWebhookVerifier(),
	}

	c.Inference = &InferenceService{client: c}
	c.Domains = &DomainsService{client: c}
	c.DNS = &DNSService{client: c}
	c.Telemetry = &TelemetryService{client: c}
	c.Compute = &ComputeService{client: c}
	c.Storage = &StorageService{client: c}
	c.Network = &NetworkService{client: c}

	return c, nil
}

// InferenceService coordinates multi-objective AI routing and completion.
type InferenceService struct {
	client *Client
}

// CreateChatCompletion executes an LLM chat completion with intelligent edge routing.
func (s *InferenceService) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req == nil {
		return nil, errors.New("aeitron: ChatCompletionRequest cannot be nil")
	}
	if req.Model == "" {
		return nil, errors.New("aeitron: ChatCompletionRequest.Model is required")
	}

	var resp ChatCompletionResponse
	if err := s.client.transport.do(ctx, "POST", "/v1/chat/completions", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DomainsService manages custom edge domains, certificates, and origin pools.
type DomainsService struct {
	client *Client
}

// List returns all onboarded domains for the current project.
func (s *DomainsService) List(ctx context.Context) ([]DomainRecord, error) {
	var resp struct {
		Domains []DomainRecord `json:"domains"`
	}
	if err := s.client.transport.do(ctx, "GET", "/v1/domains", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Domains, nil
}

// Onboard binds a new domain to the global edge network.
func (s *DomainsService) Onboard(ctx context.Context, req *OnboardDomainRequest) (*DomainRecord, error) {
	if req == nil {
		return nil, errors.New("aeitron: OnboardDomainRequest cannot be nil")
	}
	var resp DomainRecord
	if err := s.client.transport.do(ctx, "POST", "/v1/domains", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Verify triggers immediate edge DNS propagation verification.
func (s *DomainsService) Verify(ctx context.Context, domainID string) (*VerifyDomainResult, error) {
	if domainID == "" {
		return nil, errors.New("aeitron: domainID is required")
	}
	path := fmt.Sprintf("/v1/domains/%s/verify", domainID)
	var resp VerifyDomainResult
	if err := s.client.transport.do(ctx, "POST", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DNSService interacts with Anycast DNS zones and records.
type DNSService struct {
	client *Client
}

// ListZones retrieves all authoritative DNS zones.
func (s *DNSService) ListZones(ctx context.Context) ([]DNSZone, error) {
	var resp struct {
		Zones []DNSZone `json:"zones"`
	}
	if err := s.client.transport.do(ctx, "GET", "/v1/dns/zones", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Zones, nil
}

// CreateRecord adds a new resource record to a zone.
func (s *DNSService) CreateRecord(ctx context.Context, req *CreateDNSRecordRequest) (*DNSRecord, error) {
	if req == nil || req.ZoneID == "" {
		return nil, errors.New("aeitron: CreateDNSRecordRequest with ZoneID is required")
	}
	path := fmt.Sprintf("/v1/dns/zones/%s/records", req.ZoneID)
	var resp DNSRecord
	if err := s.client.transport.do(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TelemetryService accesses global real-time edge performance metrics.
type TelemetryService struct {
	client *Client
}

// GetSummary returns aggregate platform SLA telemetry.
func (s *TelemetryService) GetSummary(ctx context.Context) (*TelemetrySummary, error) {
	var resp TelemetrySummary
	if err := s.client.transport.do(ctx, "GET", "/v1/telemetry/summary", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPoPHealth returns health, latency, and status for all active edge PoPs.
func (s *TelemetryService) GetPoPHealth(ctx context.Context) ([]PoPHealthRecord, error) {
	var resp struct {
		PoPs []PoPHealthRecord `json:"pops"`
	}
	if err := s.client.transport.do(ctx, "GET", "/v1/status/pops", nil, &resp); err != nil {
		return nil, err
	}
	return resp.PoPs, nil
}

// ComputeService provisions and inspects edge microVM instances.
type ComputeService struct {
	client *Client
}

// ListInstances returns running microVM instances.
func (s *ComputeService) ListInstances(ctx context.Context) ([]MicroVMInstance, error) {
	var resp struct {
		Instances []MicroVMInstance `json:"instances"`
	}
	if err := s.client.transport.do(ctx, "GET", "/v1/compute/instances", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Instances, nil
}

// StorageService manages distributed edge object buckets.
type StorageService struct {
	client *Client
}

// ListBuckets returns all buckets in the tenant namespace.
func (s *StorageService) ListBuckets(ctx context.Context) ([]BucketRecord, error) {
	var resp struct {
		Buckets []BucketRecord `json:"buckets"`
	}
	if err := s.client.transport.do(ctx, "GET", "/v1/storage/buckets", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Buckets, nil
}

// NetworkService queries subsea routes and edge WAN latency matrices.
type NetworkService struct {
	client *Client
}

// GetLatencyMatrix returns cross-PoP WAN latency measurements.
func (s *NetworkService) GetLatencyMatrix(ctx context.Context) ([]LatencyMatrixResult, error) {
	var resp struct {
		Matrix []LatencyMatrixResult `json:"matrix"`
	}
	if err := s.client.transport.do(ctx, "GET", "/v1/network/latency-matrix", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Matrix, nil
}

