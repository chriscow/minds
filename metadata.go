package minds

// MergeStrategy defines how to handle metadata key conflicts
type MergeStrategy int

const (
	// KeepExisting keeps the existing value on conflict
	KeepExisting MergeStrategy = iota
	// KeepNew overwrites with the new value on conflict
	KeepNew
	// Combine attempts to combine values (slice/map/string)
	Combine
	// Skip ignores conflicting keys
	Skip
)

type Metadata map[string]any

// Copy creates a deep copy of the metadata
func (m Metadata) Copy() Metadata {
	copied := make(Metadata, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

// Merge combines the current metadata with another, using the specified strategy
func (m Metadata) Merge(other Metadata, strategy MergeStrategy) Metadata {
	return m.MergeWithCustom(other, strategy, nil)
}

// MergeWithCustom combines metadata with custom handlers for specific keys
func (m Metadata) MergeWithCustom(other Metadata, strategy MergeStrategy,
	customMerge map[string]func(existing, new any) any) Metadata {

	result := m.Copy()

	for k, newVal := range other {
		// Check for custom merge function first
		if customMerge != nil {
			if mergeFn, exists := customMerge[k]; exists {
				if existingVal, hasKey := result[k]; hasKey {
					result[k] = mergeFn(existingVal, newVal)
				} else {
					result[k] = newVal
				}
				continue
			}
		}

		// Handle based on strategy if key exists
		if existingVal, exists := result[k]; exists {
			switch strategy {
			case KeepExisting:
				continue
			case KeepNew:
				result[k] = newVal
			case Combine:
				result[k] = combineValues(existingVal, newVal)
			case Skip:
				continue
			}
		} else {
			// No conflict, just add the new value
			result[k] = newVal
		}
	}

	return result
}

func combineValues(existing, new any) any {
	switch existingVal := existing.(type) {
	case []any:
		if newVal, ok := new.([]any); ok {
			combined := make([]any, len(existingVal))
			copy(combined, existingVal)
			return append(combined, newVal...)
		}

	case map[string]any:
		if newVal, ok := new.(map[string]any); ok {
			combined := make(map[string]any)
			for k, v := range existingVal {
				combined[k] = v
			}
			for k, v := range newVal {
				combined[k] = v
			}
			return combined
		}

	case string:
		if newVal, ok := new.(string); ok {
			return existingVal + newVal
		}

	case int:
		if newVal, ok := new.(int); ok {
			return existingVal + newVal
		}

	case float64:
		if newVal, ok := new.(float64); ok {
			return existingVal + newVal
		}
	}

	// If types don't match or aren't combinable, return new value
	return new
}
