package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"BotMatrix/common/types"

	"github.com/gorilla/websocket"
)

// JSONResponse is an alias for types.ApiResponse used for Swagger documentation
type JSONResponse types.ApiResponse

// GenerateRandomToken generates a random hex token
func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// Upgrader for WebSocket connections
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin
	},
}

// ReadJSONWithNumber reads JSON from WebSocket and uses json.Number for precision
func ReadJSONWithNumber(conn *websocket.Conn, v any) error {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(message))
	decoder.UseNumber()
	return decoder.Decode(v)
}

// DecodeMapToStruct decodes a map into a struct using JSON
func DecodeMapToStruct(m any, v any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// SendJSONResponse sends a standard API response
func SendJSONResponse(w http.ResponseWriter, success bool, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.ApiResponse{
		Success: success,
		Message: message,
		Data:    data,
	})
}

// SendJSONResponseWithCode sends a standard API response with a custom code
func SendJSONResponseWithCode(w http.ResponseWriter, success bool, message string, code string, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.ApiResponse{
		Success: success,
		Message: message,
		Code:    code,
		Data:    data,
	})
}

// MatchRoutePattern checks if a message matches a routing pattern (regex)
func MatchRoutePattern(pattern, message string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	// Try as regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fallback to simple contains
		return strings.Contains(message, pattern)
	}
	return re.MatchString(message)
}

// ToString converts any value to string
func ToString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		// 避免科学计数法
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case json.Number:
		// 优先尝试作为整数解析，避免科学计数法和精度丢失
		if i, err := val.Int64(); err == nil {
			return strconv.FormatInt(i, 10)
		}
		// 如果是浮点数，也尝试避免科学计数法
		if f, err := val.Float64(); err == nil {
			return strconv.FormatFloat(f, 'f', -1, 64)
		}
		return val.String()
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToBool converts any value to bool
func ToBool(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int, int32, int64:
		return ToInt64(v) != 0
	case float32, float64:
		return val != 0
	case string:
		s := strings.ToLower(val)
		return s == "true" || s == "1" || s == "yes" || s == "on"
	default:
		return false
	}
}

// ToInt64 converts any value to int64
func ToInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case float32:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		var i int64
		fmt.Sscanf(val, "%d", &i)
		return i
	default:
		return 0
	}
}

// ContainsOne checks if string s contains any of the keywords
func ContainsOne(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

// ToPascalMap converts a struct to a map with PascalCase keys (matching struct field names)
// This is useful for compatibility with C# systems that expect PascalCase JSON.
func ToPascalMap(v any) map[string]any {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil
	}

	res := make(map[string]any)
	t := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}

		// Skip fields with gorm:"-" or those marked as ignored in some way if needed
		// For now, we just take all exported fields
		val := rv.Field(i).Interface()

		// Special handling for time.Time to match C# default JSON format (ISO 8601)
		if t, ok := val.(time.Time); ok {
			if t.IsZero() {
				res[field.Name] = "0001-01-01T00:00:00"
			} else {
				res[field.Name] = t.Format("2006-01-02T15:04:05")
			}
		} else {
			res[field.Name] = val
		}
	}

	return res
}
