package logreader

// ParseProfile extracts player profile information from log entries.
// It looks for authenticateResponse events that contain screenName and clientId.
func ParseProfile(entries []*LogEntry) (*PlayerProfile, error) {
	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Check if this is an authenticateResponse
		if authResp, ok := entry.JSON["authenticateResponse"]; ok {
			authMap, ok := authResp.(map[string]interface{})
			if !ok {
				continue
			}

			profile := &PlayerProfile{}
			if screenName, ok := authMap["screenName"].(string); ok {
				profile.ScreenName = screenName
			}
			if clientID, ok := authMap["clientId"].(string); ok {
				profile.ClientID = clientID
			}

			if profile.ScreenName != "" || profile.ClientID != "" {
				return profile, nil
			}
		}
	}

	return nil, nil
}

// ParseInventory extracts player inventory information from log entries.
// It looks for InventoryInfo events.
func ParseInventory(entries []*LogEntry) (*PlayerInventory, error) {
	// Look for the most recent InventoryInfo
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if !entry.IsJSON {
			continue
		}

		// Check if this is an InventoryInfo event
		if invInfo, ok := entry.JSON["InventoryInfo"]; ok {
			invMap, ok := invInfo.(map[string]interface{})
			if !ok {
				continue
			}

			inventory := &PlayerInventory{
				CustomTokens: make(map[string]int),
			}

			// Extract basic resources
			if gems, ok := invMap["Gems"].(float64); ok {
				inventory.Gems = int(gems)
			}
			if gold, ok := invMap["Gold"].(float64); ok {
				inventory.Gold = int(gold)
			}
			if vaultProgress, ok := invMap["TotalVaultProgress"].(float64); ok {
				inventory.TotalVaultProgress = int(vaultProgress)
			}

			// Extract wildcards
			if wcCommon, ok := invMap["WildCardCommons"].(float64); ok {
				inventory.WildCardCommons = int(wcCommon)
			}
			if wcUncommon, ok := invMap["WildCardUnCommons"].(float64); ok {
				inventory.WildCardUncommons = int(wcUncommon)
			}
			if wcRare, ok := invMap["WildCardRares"].(float64); ok {
				inventory.WildCardRares = int(wcRare)
			}
			if wcMythic, ok := invMap["WildCardMythics"].(float64); ok {
				inventory.WildCardMythics = int(wcMythic)
			}

			// Extract boosters
			if boosters, ok := invMap["Boosters"].([]interface{}); ok {
				for _, b := range boosters {
					boosterMap, ok := b.(map[string]interface{})
					if !ok {
						continue
					}

					booster := Booster{}
					if setCode, ok := boosterMap["SetCode"].(string); ok {
						booster.SetCode = setCode
					}
					if count, ok := boosterMap["Count"].(float64); ok {
						booster.Count = int(count)
					}
					if collationID, ok := boosterMap["CollationId"].(float64); ok {
						booster.CollationID = int(collationID)
					}

					if booster.SetCode != "" && booster.Count > 0 {
						inventory.Boosters = append(inventory.Boosters, booster)
					}
				}
			}

			// Extract custom tokens
			if tokens, ok := invMap["CustomTokens"].(map[string]interface{}); ok {
				for key, value := range tokens {
					if count, ok := value.(float64); ok {
						inventory.CustomTokens[key] = int(count)
					}
				}
			}

			return inventory, nil
		}
	}

	return nil, nil
}

// ParseRank extracts player rank information from log entries.
// It looks for JSON entries containing rank information (constructedSeasonOrdinal or limitedSeasonOrdinal).
func ParseRank(entries []*LogEntry) (*PlayerRank, error) {
	// Look for the most recent rank info
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if !entry.IsJSON {
			continue
		}

		// Check if this entry has rank information
		// Rank responses contain either constructedSeasonOrdinal or limitedSeasonOrdinal
		_, hasConstructed := entry.JSON["constructedSeasonOrdinal"]
		_, hasLimited := entry.JSON["limitedSeasonOrdinal"]

		if !hasConstructed && !hasLimited {
			continue
		}

		rank := &PlayerRank{}

		// Extract constructed rank info
		if season, ok := entry.JSON["constructedSeasonOrdinal"].(float64); ok {
			rank.ConstructedSeasonOrdinal = int(season)
		}
		if class, ok := entry.JSON["constructedClass"].(string); ok {
			rank.ConstructedClass = class
		}
		if level, ok := entry.JSON["constructedLevel"].(float64); ok {
			rank.ConstructedLevel = int(level)
		}
		if percentile, ok := entry.JSON["constructedPercentile"].(float64); ok {
			rank.ConstructedPercentile = percentile
		}
		if step, ok := entry.JSON["constructedStep"].(float64); ok {
			rank.ConstructedStep = int(step)
		}

		// Extract limited rank info
		if season, ok := entry.JSON["limitedSeasonOrdinal"].(float64); ok {
			rank.LimitedSeasonOrdinal = int(season)
		}
		if class, ok := entry.JSON["limitedClass"].(string); ok {
			rank.LimitedClass = class
		}
		if level, ok := entry.JSON["limitedLevel"].(float64); ok {
			rank.LimitedLevel = int(level)
		}
		if percentile, ok := entry.JSON["limitedPercentile"].(float64); ok {
			rank.LimitedPercentile = percentile
		}
		if step, ok := entry.JSON["limitedStep"].(float64); ok {
			rank.LimitedStep = int(step)
		}

		// Extract match statistics
		if won, ok := entry.JSON["limitedMatchesWon"].(float64); ok {
			rank.LimitedMatchesWon = int(won)
		}
		if lost, ok := entry.JSON["limitedMatchesLost"].(float64); ok {
			rank.LimitedMatchesLost = int(lost)
		}

		return rank, nil
	}

	return nil, nil
}

// ParseDraftHistory extracts draft/limited event history from log entries.
// It looks for "Courses" arrays and filters for limited format events.
func ParseDraftHistory(entries []*LogEntry) (*DraftHistory, error) {
	history := &DraftHistory{
		Drafts: []DraftEvent{},
	}

	// Track unique course IDs to avoid duplicates
	seenCourses := make(map[string]bool)

	// Look for Courses array (search from most recent)
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if !entry.IsJSON {
			continue
		}

		// Check if this entry has a Courses array
		if coursesData, ok := entry.JSON["Courses"]; ok {
			courses, ok := coursesData.([]interface{})
			if !ok {
				continue
			}

			// Process each course
			for _, courseData := range courses {
				courseMap, ok := courseData.(map[string]interface{})
				if !ok {
					continue
				}

				// Extract course ID
				courseID, _ := courseMap["CourseId"].(string)
				if courseID == "" || seenCourses[courseID] {
					continue
				}

				// Extract event name
				eventName, _ := courseMap["InternalEventName"].(string)

				// Filter for limited/draft events
				// Look for keywords: Draft, Sealed, Premier, Quick, Traditional
				isDraftEvent := false
				lowerEvent := ""
				if eventName != "" {
					lowerEvent = eventName
					if contains(lowerEvent, "Draft") ||
						contains(lowerEvent, "Sealed") ||
						contains(lowerEvent, "Premier") ||
						contains(lowerEvent, "Quick") ||
						contains(lowerEvent, "Traditional") {
						isDraftEvent = true
					}
				}

				if !isDraftEvent {
					continue
				}

				// Mark as seen
				seenCourses[courseID] = true

				// Create draft event
				draftEvent := DraftEvent{
					EventID:   courseID,
					EventName: eventName,
				}

				// Extract status/module
				if status, ok := courseMap["CurrentModule"].(string); ok {
					draftEvent.Status = status
				}

				// Extract wins
				if wins, ok := courseMap["CurrentWins"].(float64); ok {
					draftEvent.Wins = int(wins)
				}

				// Extract losses if available
				if losses, ok := courseMap["CurrentLosses"].(float64); ok {
					draftEvent.Losses = int(losses)
				}

				// Extract deck information
				if courseDeckData, ok := courseMap["CourseDeck"]; ok {
					if deckMap, ok := courseDeckData.(map[string]interface{}); ok {
						draftEvent.Deck = parseDraftDeck(deckMap)
					}
				}

				// Extract deck summary for name
				if deckSummary, ok := courseMap["CourseDeckSummary"]; ok {
					if summaryMap, ok := deckSummary.(map[string]interface{}); ok {
						if name, ok := summaryMap["Name"].(string); ok {
							draftEvent.Deck.Name = name
						}
					}
				}

				history.Drafts = append(history.Drafts, draftEvent)
			}

			// We found a Courses array, no need to continue
			break
		}
	}

	if len(history.Drafts) == 0 {
		return nil, nil
	}

	return history, nil
}

// parseDraftDeck extracts deck information from a CourseDeck object.
func parseDraftDeck(deckMap map[string]interface{}) DraftDeck {
	deck := DraftDeck{
		MainDeck: []DeckCard{},
	}

	// Extract MainDeck
	if mainDeckData, ok := deckMap["MainDeck"]; ok {
		if mainDeck, ok := mainDeckData.([]interface{}); ok {
			for _, cardData := range mainDeck {
				if cardMap, ok := cardData.(map[string]interface{}); ok {
					card := DeckCard{}

					if cardID, ok := cardMap["cardId"].(float64); ok {
						card.CardID = int(cardID)
					}

					if quantity, ok := cardMap["quantity"].(float64); ok {
						card.Quantity = int(quantity)
					}

					if card.CardID > 0 && card.Quantity > 0 {
						deck.MainDeck = append(deck.MainDeck, card)
					}
				}
			}
		}
	}

	return deck
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	// Check if substr appears in s
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// ParseAll extracts all available information from log entries.
func ParseAll(entries []*LogEntry) (*PlayerProfile, *PlayerInventory, *PlayerRank) {
	profile, _ := ParseProfile(entries)
	inventory, _ := ParseInventory(entries)
	rank, _ := ParseRank(entries)

	return profile, inventory, rank
}
