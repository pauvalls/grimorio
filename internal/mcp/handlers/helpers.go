package handlers

import "strconv"

// getStringArg extracts a string argument from the args map
func getStringArg(args map[string]any, key string) string {
	if val, ok := args[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// getIntArg extracts an int argument from the args map
func getIntArg(args map[string]any, key string) int {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// getInt64Arg extracts an int64 argument from the args map
func getInt64Arg(args map[string]any, key string) int64 {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int:
			return int64(v)
		case int64:
			return v
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	}
	return 0
}

// getArrayArg extracts an array argument from the args map
func getArrayArg(args map[string]any, key string) []any {
	if val, ok := args[key]; ok {
		if arr, ok := val.([]any); ok {
			return arr
		}
	}
	return nil
}
