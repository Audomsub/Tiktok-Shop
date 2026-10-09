package usecase_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type workflowDefinition struct {
	Name        string                            `json:"name"`
	Nodes       []workflowNode                    `json:"nodes"`
	Connections map[string]map[string][][]connectionTarget `json:"connections"`
}

type workflowNode struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	OnError    string                 `json:"onError,omitempty"`
}

type connectionTarget struct {
	Node  string `json:"node"`
	Type  string `json:"type"`
	Index int    `json:"index"`
}

func TestFastMossScraperWorkflow_TemplateIntegrity(t *testing.T) {
	// Locate workflow relative to current package directory
	workflowPath := filepath.Join("..", "..", "..", "n8n", "workflows", "fastmoss_scraper_workflow.json")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read workflow JSON file at %s: %v", workflowPath, err)
	}

	var wf workflowDefinition
	if err := json.Unmarshal(data, &wf); err != nil {
		t.Fatalf("Workflow JSON is invalid: %v", err)
	}

	if wf.Name == "" {
		t.Errorf("Workflow name should not be empty")
	}

	nodesByName := make(map[string]workflowNode)
	for _, n := range wf.Nodes {
		if _, exists := nodesByName[n.Name]; exists {
			t.Errorf("Duplicate node name found: %s", n.Name)
		}
		nodesByName[n.Name] = n
	}

	// 1. Verify Schedule Triggers
	shallow, exists := nodesByName["Schedule Trigger - Shallow Crawl (06:00, 12:00, 18:00)"]
	if !exists {
		t.Errorf("Missing shallow crawl schedule trigger node")
	} else {
		rule, _ := shallow.Parameters["rule"].(map[string]interface{})
		interval, _ := rule["interval"].([]interface{})
		if len(interval) > 0 {
			item, _ := interval[0].(map[string]interface{})
			cronExp, _ := item["expression"].(string)
			if cronExp != "0 6,12,18 * * *" {
				t.Errorf("Expected shallow cron '0 6,12,18 * * *', got '%s'", cronExp)
			}
		} else {
			t.Errorf("Shallow crawl interval is empty")
		}
	}

	deep, exists := nodesByName["Schedule Trigger - Deep Crawl (00:00 Midnight)"]
	if !exists {
		t.Errorf("Missing deep crawl schedule trigger node")
	} else {
		rule, _ := deep.Parameters["rule"].(map[string]interface{})
		interval, _ := rule["interval"].([]interface{})
		if len(interval) > 0 {
			item, _ := interval[0].(map[string]interface{})
			cronExp, _ := item["expression"].(string)
			if cronExp != "0 0 * * *" {
				t.Errorf("Expected deep cron '0 0 * * *', got '%s'", cronExp)
			}
		} else {
			t.Errorf("Deep crawl interval is empty")
		}
	}

	// 2. Verify Cookie Injection in FastMoss Request Node
	fetchNode, exists := nodesByName["HTTP Request - Fetch FastMoss Page"]
	if !exists {
		t.Errorf("Missing FastMoss HTTP request node")
	} else {
		headerParams, _ := fetchNode.Parameters["headerParameters"].(map[string]interface{})
		params, _ := headerParams["parameters"].([]interface{})
		hasCookie := false
		for _, p := range params {
			pMap, _ := p.(map[string]interface{})
			if strings.EqualFold(pMap["name"].(string), "Cookie") {
				val, _ := pMap["value"].(string)
				if strings.Contains(val, "$env.FASTMOSS_COOKIE") {
					hasCookie = true
				}
			}
		}
		if !hasCookie {
			t.Errorf("HTTP Request node does not properly inject $env.FASTMOSS_COOKIE")
		}
	}

	// 3. Verify Anti-Ban Jitter and Cool-Off logic
	jitterNode, exists := nodesByName["Calculate Jitter Delay & Cool-Off Pause"]
	if !exists {
		t.Errorf("Missing Jitter & Cool-off calculation node")
	} else {
		jsCode, _ := jitterNode.Parameters["jsCode"].(string)
		if !strings.Contains(jsCode, "7.0 - 3.0") || !strings.Contains(jsCode, "coolOffSeconds = 25") {
			t.Errorf("Jitter calculation does not implement 3-7s randomized jitter or 25s cool-off pause")
		}
	}

	// 4. Verify Immediate Halting on HTTP 403 / 429
	stopNode, exists := nodesByName["Stop and Terminate Workflow Immediately"]
	if !exists {
		t.Errorf("Missing Stop and Terminate node")
	} else if stopNode.Type != "n8n-nodes-base.stopAndError" {
		t.Errorf("Expected stop node type 'n8n-nodes-base.stopAndError', got '%s'", stopNode.Type)
	}

	// 5. Verify Workflow Topology Integrity
	for source, targets := range wf.Connections {
		if _, ok := nodesByName[source]; !ok {
			t.Errorf("Connection source '%s' does not exist in nodes", source)
		}
		for _, outputList := range targets {
			for _, connList := range outputList {
				for _, conn := range connList {
					if _, ok := nodesByName[conn.Node]; !ok {
						t.Errorf("Connection target '%s' from '%s' does not exist in nodes", conn.Node, source)
					}
				}
			}
		}
	}
}
