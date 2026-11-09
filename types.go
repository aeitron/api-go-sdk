package aeitron

import "time"

// RoutingStrategy defines multi-objective AI routing policies.
type RoutingStrategy string

const (
	RoutingLatencyOptimal RoutingStrategy = "LATENCY_OPTIMAL"
	RoutingCostMinimal    RoutingStrategy = "COST_MINIMAL"
	RoutingCarbonNeutral  RoutingStrategy = "CARBON_NEUTRAL"
	RoutingSovereignPinned RoutingStrategy = "SOVEREIGN_PINNED"
)

// Jurisdiction represents geopolitical boundary pinning.
type Jurisdiction string

const (
	JurisdictionBD     Jurisdiction = "BD"
	JurisdictionEU     Jurisdiction = "EU"
	JurisdictionUS     Jurisdiction = "US"
	JurisdictionGlobal Jurisdiction = "GLOBAL"
)

// ChatMessage represents an LLM conversation turn.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest is the payload for AI model inference.
type ChatCompletionRequest struct {
	Model           string          `json:"model"`
	Messages        []ChatMessage   `json:"messages"`
	RoutingStrategy RoutingStrategy `json:"routing_strategy,omitempty"`
	Jurisdiction    Jurisdiction    `json:"jurisdiction,omitempty"`
	Temperature     float64         `json:"temperature,omitempty"`
	MaxTokens       int             `json:"max_tokens,omitempty"`
}

// CompletionChoice represents an individual completion candidate.
type CompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

// CompletionTelemetry exposes edge metrics for the completion.
type CompletionTelemetry struct {
	Provider string  `json:"provider,omitempty"`
	Cost     float64 `json:"cost,omitempty"`
	TTFTMs   float64 `json:"ttft_ms,omitempty"`
	Region   string  `json:"region,omitempty"`
}

// CompletionUsage details token consumption.
type CompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse is returned by the AI inference gateway.
type ChatCompletionResponse struct {
	ID        string              `json:"id,omitempty"`
	Object    string              `json:"object,omitempty"`
	Created   int64               `json:"created,omitempty"`
	Model     string              `json:"model"`
	Choices   []CompletionChoice  `json:"choices"`
	Usage     CompletionUsage     `json:"usage"`
	Telemetry CompletionTelemetry `json:"telemetry"`
}

// DomainRecord represents an onboarded custom edge domain.
type DomainRecord struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	Hostname       string    `json:"hostname"`
	CNAMETarget    string    `json:"cname_target,omitempty"`
	OriginAddress  string    `json:"origin_address,omitempty"`
	OriginPort     int       `json:"origin_port,omitempty"`
	OriginProtocol string    `json:"origin_protocol,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}

// OnboardDomainRequest defines properties to bind a domain to the edge.
type OnboardDomainRequest struct {
	ProjectID      string `json:"project_id"`
	Hostname       string `json:"hostname"`
	OriginAddress  string `json:"origin_address"`
	OriginPort     int    `json:"origin_port"`
	OriginProtocol string `json:"origin_protocol,omitempty"`
}

// VerifyDomainResult returns status after checking DNS verification records.
type VerifyDomainResult struct {
	Verified bool   `json:"verified"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
}

// DNSZone represents an authoritative DNS zone.
type DNSZone struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	TTL       int       `json:"ttl"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// DNSRecord represents a resource record within a DNS zone.
type DNSRecord struct {
	ID      string `json:"id"`
	ZoneID  string `json:"zone_id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// CreateDNSRecordRequest specifies fields for adding a DNS record.
type CreateDNSRecordRequest struct {
	ZoneID  string `json:"zone_id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl,omitempty"`
}

// TelemetrySummary provides aggregate global platform SLA metrics.
type TelemetrySummary struct {
	GlobalRequestsPerSec float64 `json:"global_requests_per_sec"`
	EdgeCacheHitRatio    float64 `json:"edge_cache_hit_ratio"`
	P99LatencyMs         float64 `json:"p99_latency_ms"`
	TotalBlockedAttacks  int64   `json:"total_blocked_attacks"`
}

// PoPHealthRecord details the health of a specific edge point of presence.
type PoPHealthRecord struct {
	PoPID      string  `json:"pop_id"`
	Metro      string  `json:"metro"`
	Country    string  `json:"country"`
	Status     string  `json:"status"`
	LatencyMs  float64 `json:"latency_ms"`
	PacketLoss float64 `json:"packet_loss"`
}

// ComputeNodeRecord describes an edge bare-metal or microVM compute host.
type ComputeNodeRecord struct {
	NodeID          string  `json:"node_id"`
	Metro           string  `json:"metro"`
	Status          string  `json:"status"`
	CPUUsagePercent float64 `json:"cpu_usage_percent"`
	MemUsagePercent float64 `json:"mem_usage_percent"`
}

// MicroVMInstance represents an ephemeral microVM worker.
type MicroVMInstance struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	VCPUs     int    `json:"vcpus"`
	MemoryMB  int    `json:"memory_mb"`
	Status    string `json:"status"`
	IPAddress string `json:"ip_address,omitempty"`
}

// BucketRecord represents an object storage bucket.
type BucketRecord struct {
	Name      string    `json:"name"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
}

// ObjectRecord represents a single object stored in edge storage.
type ObjectRecord struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ContentType  string    `json:"content_type"`
}

// LatencyMatrixResult measures WAN propagation between PoP coordinates.
type LatencyMatrixResult struct {
	SourcePoP string  `json:"source_pop"`
	DestPoP   string  `json:"dest_pop"`
	LatencyMs float64 `json:"latency_ms"`
	JitterMs  float64 `json:"jitter_ms"`
}

// SubseaCableStatus details international transit path health.
type SubseaCableStatus struct {
	CableID      string  `json:"cable_id"`
	Name         string  `json:"name"`
	Operational  bool    `json:"operational"`
	CapacityTbps float64 `json:"capacity_tbps"`
}

