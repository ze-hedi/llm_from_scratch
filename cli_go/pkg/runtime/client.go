package runtime

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client communicates with the otto_code runtime server and backend API.
type Client struct {
	BaseURL    string // Runtime server (e.g. http://localhost:5000)
	APIURL     string // Backend API (e.g. http://localhost:4000)
	HTTPClient *http.Client
}

// AgentData describes an agent for the /runtime/run endpoint.
type AgentData struct {
	ID            string `json:"_id"`
	Name          string `json:"name"`
	Model         string `json:"model"`
	Description   string `json:"description,omitempty"`
	ThinkingLevel string `json:"thinkingLevel,omitempty"`
	SessionMode   string `json:"sessionMode,omitempty"`
	WorkingDir    string `json:"workingDir,omitempty"`
	Playground    string `json:"playground,omitempty"`
	APIKey        string `json:"apiKey,omitempty"`
	Stateful      bool   `json:"stateful,omitempty"`
}

// FilePayload carries a soul or skills file for agent initialization.
type FilePayload struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// RunRequest is the body for POST /runtime/run.
type RunRequest struct {
	Agent     AgentData     `json:"agent"`
	Files     []FilePayload `json:"files,omitempty"`
	SessionID string        `json:"sessionId,omitempty"`
}

// RunResponse is returned by POST /runtime/run.
type RunResponse struct {
	Success         bool   `json:"success"`
	AgentID         string `json:"agentId"`
	SessionID       string `json:"sessionId"`
	Name            string `json:"name"`
	Model           string `json:"model"`
	SessionMode     string `json:"sessionMode"`
	ThinkingLevel   string `json:"thinkingLevel"`
	WorkingDir      string `json:"workingDir"`
	HasCustomAPIKey bool   `json:"hasCustomApiKey"`
}

// StatusResponse is returned by GET /runtime/status.
type StatusResponse struct {
	ActiveAgents    []string          `json:"activeAgents"`
	SessionAgentMap map[string]string `json:"sessionAgentMap"`
	CurrentAgentID  string            `json:"currentAgentId"`
}

// SubAgentEntry represents one sub-agent inside an orchestrator (from /api/orchestrators).
type SubAgentEntry struct {
	Agent    AgentData `json:"agent"`
	Stateful bool      `json:"stateful"`
}

// OrchestratorData describes an orchestrator from the backend API.
type OrchestratorData struct {
	ID          string          `json:"_id"`
	Name        string          `json:"name"`
	Model       string          `json:"model"`
	Description string          `json:"description,omitempty"`
	Playground  string          `json:"playground,omitempty"`
	SubAgents   []SubAgentEntry `json:"subAgents"`
}

// OrchestratorRunRequest is the body for POST /runtime/orchestrator/run.
type OrchestratorRunRequest struct {
	OrchestratorID string      `json:"orchestratorId"`
	SessionID      string      `json:"sessionId,omitempty"`
	SystemPrompt   string      `json:"systemPrompt"`
	Model          string      `json:"model,omitempty"`
	Playground     string      `json:"playground,omitempty"`
	Agents         []AgentData `json:"agents"`
}

// OrchestratorRunResponse is returned by POST /runtime/orchestrator/run.
type OrchestratorRunResponse struct {
	Success        bool     `json:"success"`
	OrchestratorID string   `json:"orchestratorId"`
	SessionID      string   `json:"sessionId"`
	Model          string   `json:"model"`
	SubAgents      []string `json:"subAgents"`
}

// SSEEvent represents a single parsed Server-Sent Event from the chat stream.
type SSEEvent struct {
	Type    string          `json:"type"`
	Text    string          `json:"text,omitempty"`
	Name    string          `json:"name,omitempty"`
	Args    json.RawMessage `json:"args,omitempty"`
	Result  string          `json:"result,omitempty"`
	IsError bool            `json:"isError,omitempty"`
	Message string          `json:"message,omitempty"`
}

func NewClient(baseURL, apiURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIURL:  apiURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// ListAgents fetches all available agents from the backend API.
func (c *Client) ListAgents() ([]AgentData, error) {
	resp, err := c.HTTPClient.Get(c.APIURL + "/api/agents")
	if err != nil {
		return nil, fmt.Errorf("connect to API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var agents []AgentData
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, fmt.Errorf("parse agents response: %w", err)
	}
	return agents, nil
}

// Run creates or re-initializes an agent session via POST /runtime/run.
func (c *Client) Run(req RunRequest) (*RunResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal run request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/runtime/run", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create run request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("connect to runtime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("runtime error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse run response: %w", err)
	}
	return &result, nil
}

// ChatStream sends a message and returns a channel of SSE events streamed back.
// The channel is closed when the stream ends (done/error/EOF).
// The caller must consume the channel to avoid leaking the reading goroutine.
func (c *Client) ChatStream(sessionID, message string) (<-chan SSEEvent, error) {
	body, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/runtime/chat/"+sessionID, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create chat request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// Use a client without timeout for streaming — the stream can last a long time.
	streamClient := &http.Client{}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("connect to runtime: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("runtime error (status %d): %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan SSEEvent, 16)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			raw := strings.TrimPrefix(line, "data: ")

			var event SSEEvent
			if err := json.Unmarshal([]byte(raw), &event); err != nil {
				continue
			}

			ch <- event

			if event.Type == "done" {
				return
			}
		}
	}()

	return ch, nil
}

// GetStatus returns the list of active agents and session mappings.
func (c *Client) GetStatus() (*StatusResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/runtime/status")
	if err != nil {
		return nil, fmt.Errorf("connect to runtime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("runtime error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse status response: %w", err)
	}
	return &result, nil
}

// AbortAgent aborts the active session for the given agent.
func (c *Client) AbortAgent(sessionID string) error {
	httpReq, err := http.NewRequest("POST", c.BaseURL+"/runtime/agents/"+sessionID+"/abort", nil)
	if err != nil {
		return fmt.Errorf("create abort request: %w", err)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("connect to runtime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("runtime error (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// ListOrchestrators fetches all available orchestrators from the backend API.
func (c *Client) ListOrchestrators() ([]OrchestratorData, error) {
	resp, err := c.HTTPClient.Get(c.APIURL + "/api/orchestrators")
	if err != nil {
		return nil, fmt.Errorf("connect to API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var orchestrators []OrchestratorData
	if err := json.NewDecoder(resp.Body).Decode(&orchestrators); err != nil {
		return nil, fmt.Errorf("parse orchestrators response: %w", err)
	}
	return orchestrators, nil
}

// RunOrchestrator creates an orchestrator session via POST /runtime/orchestrator/run.
func (c *Client) RunOrchestrator(req OrchestratorRunRequest) (*OrchestratorRunResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal orchestrator run request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/runtime/orchestrator/run", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create orchestrator run request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("connect to runtime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("runtime error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result OrchestratorRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parse orchestrator run response: %w", err)
	}
	return &result, nil
}

