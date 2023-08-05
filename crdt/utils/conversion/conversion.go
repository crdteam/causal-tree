package conversion

import (
	"fmt"
	"strconv"
	"strings"
)

// ToString converts a json element to a string.
func ToString(data interface{}) string {
	switch v := data.(type) {
	case int32:
		return strconv.FormatInt(data.(int64), 10)
	case string:
		return data.(string)
	case []interface{}:
		var resultingStringBuilder strings.Builder
		for _, element := range data.([]interface{}) {
			resultingStringBuilder.WriteString(ToString(element))
		}
		return resultingStringBuilder.String()
	default:
		panic(fmt.Sprintf("ToString: invalid json element (%T)", v))
	}
}
