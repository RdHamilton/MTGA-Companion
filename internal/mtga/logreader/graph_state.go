package logreader

import (
	"time"
)

// QuestCompletion represents a completed quest with payout status.
type QuestCompletion struct {
	QuestNumber     int       // Quest number (1-7)
	QuestCompleted  bool      // Quest itself is completed
	PayoutCompleted bool      // Payout has been claimed
	Timestamp       time.Time // When this state was observed
}

// NodeState represents the status of any node in the graph.
type NodeState struct {
	NodeName string
	Status   string // "Available", "Completed", "Locked", etc.
}

// GraphStateData contains parsed data from GraphGetGraphState events.
type GraphStateData struct {
	CompletedQuests []QuestCompletion
	AllNodes        []NodeState // All node states for future use
	Timestamp       time.Time
}

// ParseGraphState parses GraphGetGraphState events to extract quest completion status.
// GraphGetGraphState events contain NodeStates with Quest1-7 and Quest1-7Payout completion.
func ParseGraphState(entries []*LogEntry) ([]*GraphStateData, error) {
	var results []*GraphStateData

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Parse timestamp
		timestamp := time.Now()
		if entry.Timestamp != "" {
			if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
				timestamp = parsedTime
			}
		}

		// Look for NodeStates in the JSON
		// GraphGetGraphState events have this structure:
		// {"NodeStates": {"Quest1": {"Status": "Completed"}, "Quest1Payout": {...}, ...}, "MilestoneStates": {...}}
		nodeStates, ok := entry.JSON["NodeStates"].(map[string]interface{})
		if !ok || nodeStates == nil {
			continue
		}

		// Parse quest completion states
		data := &GraphStateData{
			CompletedQuests: []QuestCompletion{},
			AllNodes:        []NodeState{},
			Timestamp:       timestamp,
		}

		// Capture ALL node states for future analysis
		for nodeName, nodeData := range nodeStates {
			if nodeMap, ok := nodeData.(map[string]interface{}); ok {
				if status, ok := nodeMap["Status"].(string); ok {
					data.AllNodes = append(data.AllNodes, NodeState{
						NodeName: nodeName,
						Status:   status,
					})
				}
			}
		}

		// Check Quest1-7 and their payouts
		questKeys := []string{"Quest1", "Quest2", "Quest3", "Quest4", "Quest5", "Quest6", "Quest7"}
		for i, questKey := range questKeys {
			questNum := i + 1
			payoutKey := questKey + "Payout"

			questCompleted := false
			payoutCompleted := false

			// Check quest completion
			if questNode, ok := nodeStates[questKey].(map[string]interface{}); ok {
				if status, ok := questNode["Status"].(string); ok && status == "Completed" {
					questCompleted = true
				}
			}

			// Check payout completion
			if payoutNode, ok := nodeStates[payoutKey].(map[string]interface{}); ok {
				if status, ok := payoutNode["Status"].(string); ok && status == "Completed" {
					payoutCompleted = true
				}
			}

			// Only add if at least one is completed
			if questCompleted || payoutCompleted {
				data.CompletedQuests = append(data.CompletedQuests, QuestCompletion{
					QuestNumber:     questNum,
					QuestCompleted:  questCompleted,
					PayoutCompleted: payoutCompleted,
					Timestamp:       timestamp,
				})
			}
		}

		// Add if we found any data (quests or other nodes)
		if len(data.CompletedQuests) > 0 || len(data.AllNodes) > 0 {
			results = append(results, data)
		}
	}

	return results, nil
}
