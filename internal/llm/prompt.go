package llm

import (
	"fmt"
	"strings"

	"github.com/eiladin/guardian-log/internal/storage"
)

// BuildPrompt constructs the LLM prompt for analyzing a DNS query
func BuildPrompt(query storage.DNSQuery, whois *storage.WHOISData) string {
	var sb strings.Builder

	sb.WriteString("You are a cybersecurity expert analyzing DNS queries for potential threats.\n\n")

	// Query information
	sb.WriteString("## DNS Query Details\n")
	sb.WriteString(fmt.Sprintf("- **Domain**: %s\n", query.Domain))
	sb.WriteString(fmt.Sprintf("- **Client**: %s (%s)\n", query.ClientName, query.ClientID))
	sb.WriteString(fmt.Sprintf("- **Query Type**: %s\n", query.QueryType))
	sb.WriteString(fmt.Sprintf("- **Response**: %s\n", query.Response))
	if query.Upstream != "" {
		sb.WriteString(fmt.Sprintf("- **Upstream**: %s\n", query.Upstream))
	}
	sb.WriteString("\n")

	// WHOIS enrichment data
	if whois != nil {
		sb.WriteString("## Domain Information (WHOIS)\n")

		if whois.Registrar != "" {
			sb.WriteString(fmt.Sprintf("- **Registrar**: %s\n", whois.Registrar))
		}

		if whois.Country != "" {
			sb.WriteString(fmt.Sprintf("- **Country**: %s\n", whois.Country))
		}

		if whois.CreatedDate != "" {
			sb.WriteString(fmt.Sprintf("- **Created**: %s\n", whois.CreatedDate))
		}

		if whois.UpdatedDate != "" {
			sb.WriteString(fmt.Sprintf("- **Updated**: %s\n", whois.UpdatedDate))
		}

		if whois.ExpiryDate != "" {
			sb.WriteString(fmt.Sprintf("- **Expires**: %s\n", whois.ExpiryDate))
		}

		if len(whois.NameServers) > 0 {
			sb.WriteString(fmt.Sprintf("- **Name Servers**: %s\n", strings.Join(whois.NameServers, ", ")))
		}

		sb.WriteString("\n")
	}

	// Analysis instructions
	sb.WriteString("## Analysis Task\n")
	sb.WriteString("This domain was identified as a **first-time query** from this client. ")
	sb.WriteString("Analyze this DNS query for potential security threats considering:\n\n")
	sb.WriteString("1. **Domain Reputation**: Is this a known malicious domain? Does it exhibit suspicious patterns?\n")
	sb.WriteString("2. **WHOIS Patterns**: Recent registration? Privacy-protected? Unusual registrar or country?\n")
	sb.WriteString("3. **Query Context**: Does the query type match expected behavior for this domain?\n")
	sb.WriteString("4. **Infrastructure**: Are the name servers or hosting infrastructure suspicious?\n\n")

	// Response format
	sb.WriteString("## Required Response Format\n")
	sb.WriteString("Respond **only** with valid JSON in the following format (no additional text):\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"classification\": \"Safe|Suspicious|Malicious\",\n")
	sb.WriteString("  \"explanation\": \"Brief explanation of your assessment\",\n")
	sb.WriteString("  \"risk_score\": 1-10,\n")
	sb.WriteString("  \"suggested_action\": \"Allow|Investigate|Block\"\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	// Classification guidelines
	sb.WriteString("### Classification Guidelines\n")
	sb.WriteString("- **Safe** (1-3): Legitimate domain from reputable organizations\n")
	sb.WriteString("- **Suspicious** (4-7): Unusual patterns that warrant investigation\n")
	sb.WriteString("- **Malicious** (8-10): Known threats or clear indicators of malicious activity\n\n")

	// Action guidelines
	sb.WriteString("### Action Guidelines\n")
	sb.WriteString("- **Allow**: No action needed, domain appears safe\n")
	sb.WriteString("- **Investigate**: Flag for manual review, potential risk\n")
	sb.WriteString("- **Block**: Immediate threat, recommend blocking\n")

	return sb.String()
}

// BuildBatchPrompt constructs a prompt for analyzing multiple queries at once
// This can be more efficient with some LLM providers
func BuildBatchPrompt(queries []storage.DNSQuery, whoisData map[string]*storage.WHOISData) string {
	var sb strings.Builder

	sb.WriteString("Analyze these DNS queries for security threats. Respond with JSON array only.\n\n")

	for i, query := range queries {
		sb.WriteString(fmt.Sprintf("%d. %s", i+1, query.Domain))

		if whois, ok := whoisData[query.Domain]; ok && whois != nil {
			if whois.Country != "" {
				sb.WriteString(fmt.Sprintf(" [%s]", whois.Country))
			}
			if whois.Registrar != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", whois.Registrar))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nFormat: [{\"domain\":\"x.com\",\"classification\":\"Safe|Suspicious|Malicious\",\"explanation\":\"...\",\"risk_score\":1-10,\"suggested_action\":\"Allow|Investigate|Block\"}]\n")

	return sb.String()
}
