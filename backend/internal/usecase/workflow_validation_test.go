package usecase_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type workflowDefinition struct {
	Name        string                                      `json:"name"`
	Nodes       []workflowNode                              `json:"nodes"`
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

	// 1. Verify Schedule Triggers (Ticket 15)
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

	// 2. Verify Cookie Injection in FastMoss Request Node (Ticket 15)
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

	// 3. Verify Anti-Ban Jitter and Cool-Off logic (Ticket 15)
	jitterNode, exists := nodesByName["Calculate Jitter Delay & Cool-Off Pause"]
	if !exists {
		t.Errorf("Missing Jitter & Cool-off calculation node")
	} else {
		jsCode, _ := jitterNode.Parameters["jsCode"].(string)
		if !strings.Contains(jsCode, "7.0 - 3.0") || !strings.Contains(jsCode, "coolOffSeconds = 25") {
			t.Errorf("Jitter calculation does not implement 3-7s randomized jitter or 25s cool-off pause")
		}
	}

	// 4. Verify Immediate Halting on HTTP 403 / 429 (Ticket 15)
	stopNode, exists := nodesByName["Stop and Terminate Workflow Immediately"]
	if !exists {
		t.Errorf("Missing Stop and Terminate node")
	} else if stopNode.Type != "n8n-nodes-base.stopAndError" {
		t.Errorf("Expected stop node type 'n8n-nodes-base.stopAndError', got '%s'", stopNode.Type)
	}

	// 5. Verify Pre-Filtering Quality Gate Rules (Ticket 16)
	gateNode, exists := nodesByName["Pre-Filter Quality Gate & Taxonomy Extraction"]
	if !exists {
		t.Errorf("Missing Pre-Filter Quality Gate node")
	} else {
		code, _ := gateNode.Parameters["jsCode"].(string)
		if !strings.Contains(code, "commissionRate >= 10.0") ||
			!strings.Contains(code, "price >= 80.0") ||
			!strings.Contains(code, "price <= 1500.0") ||
			!strings.Contains(code, "totalSales >= 30") {
			t.Errorf("Pre-Filter node missing required criteria (commission >= 10%%, price 80-1500, sales >= 30)")
		}
	}

	// 6. Verify Supabase ON CONFLICT Upsert Nodes (Ticket 16)
	upsertCatNode, exists := nodesByName["Supabase - Upsert Categories"]
	if !exists {
		t.Errorf("Missing Supabase Upsert Categories node")
	} else {
		url, _ := upsertCatNode.Parameters["url"].(string)
		if !strings.Contains(url, "on_conflict=slug") {
			t.Errorf("Categories upsert does not have on_conflict=slug in url: %s", url)
		}
	}

	upsertProdNode, exists := nodesByName["Supabase - Upsert Products"]
	if !exists {
		t.Errorf("Missing Supabase Upsert Products node")
	} else {
		url, _ := upsertProdNode.Parameters["url"].(string)
		if !strings.Contains(url, "on_conflict=source_id") {
			t.Errorf("Products upsert does not have on_conflict=source_id in url: %s", url)
		}
	}

	insertSnapshotsNode, exists := nodesByName["Supabase - Insert Product Snapshots"]
	if !exists {
		t.Errorf("Missing Supabase Insert Product Snapshots node")
	} else {
		url, _ := insertSnapshotsNode.Parameters["url"].(string)
		if !strings.Contains(url, "on_conflict=product_id,crawl_log_id") {
			t.Errorf("Snapshots insert does not have on_conflict=product_id,crawl_log_id in url: %s", url)
		}
	}

	// 7. Verify crawl_logs Audit Counts Update (Ticket 16)
	auditNode, exists := nodesByName["Supabase - Update crawl_logs Audit Counts"]
	if !exists {
		t.Errorf("Missing Supabase Update crawl_logs Audit Counts node")
	} else {
		body, _ := auditNode.Parameters["jsonBody"].(string)
		if !strings.Contains(body, "total_pages_requested") ||
			!strings.Contains(body, "total_pages_success") ||
			!strings.Contains(body, "raw_products_scraped") ||
			!strings.Contains(body, "filtered_products_saved") {
			t.Errorf("Audit counts update missing required count fields in body: %s", body)
		}
	}

	// 8. Verify Go Analytics Engine Webhook Trigger (Ticket 16)
	webhookNode, exists := nodesByName["HTTP Request - Trigger Go Analytics Engine"]
	if !exists {
		t.Errorf("Missing HTTP Request Trigger Go Analytics Engine node")
	} else {
		url, _ := webhookNode.Parameters["url"].(string)
		if !strings.Contains(url, "/api/v1/jobs/compute-scores") {
			t.Errorf("Webhook node URL does not target /api/v1/jobs/compute-scores: %s", url)
		}

		headerParams, _ := webhookNode.Parameters["headerParameters"].(map[string]interface{})
		params, _ := headerParams["parameters"].([]interface{})
		hasAPIKey := false
		for _, p := range params {
			pMap, _ := p.(map[string]interface{})
			if pMap["name"] == "X-API-Key" {
				val, _ := pMap["value"].(string)
				if strings.Contains(val, "$env.INTERNAL_API_KEY") {
					hasAPIKey = true
				}
			}
		}
		if !hasAPIKey {
			t.Errorf("Webhook node does not pass authenticated X-API-Key header referencing $env.INTERNAL_API_KEY")
		}
	}

	// 9. Verify Workflow Topology Integrity
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
