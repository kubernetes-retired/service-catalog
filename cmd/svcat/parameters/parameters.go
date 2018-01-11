package parameters

import (
	"fmt"
	"regexp"
	"strings"
)

var keymapRegex = regexp.MustCompile(`^([^\[]+)\[(.+)\]\s*$`)

// ParseVariableAssignments converts a string array of variable assignments
// into a map of keys and values
// Example:
// [a=b c=abc1232===] becomes map[a:b c:abc1232===]
func ParseVariableAssignments(params []string) (map[string]string, error) {
	variables := map[string]string{}

	for _, p := range params {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid parameter (%s), must be in name=value format", p)
		}

		variable := strings.TrimSpace(parts[0])
		if variable == "" {
			return nil, fmt.Errorf("invalid parameter (%s), variable name is requried", p)
		}
		value := strings.TrimSpace(parts[1])

		variables[variable] = value
	}

	return variables, nil
}

// ParseKeyMaps converts a string array of key lookups
// into a map of the map name and key
// Example:
// [a[b] mysecret[foo.txt]] becomes map[a:b mysecret:foo.txt]
func ParseKeyMaps(params []string) (map[string]string, error) {
	keymap := map[string]string{}

	for _, p := range params {
		parts := keymapRegex.FindStringSubmatch(p)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid parameter (%s), must be in MAP[KEY] format", p)
		}

		mapName := strings.TrimSpace(parts[1])
		if mapName == "" {
			return nil, fmt.Errorf("invalid parameter (%s), map is requried", p)
		}

		key := strings.TrimSpace(parts[2])
		if key == "" {
			return nil, fmt.Errorf("invalid parameter (%s), key is required", p)
		}

		keymap[mapName] = key
	}

	return keymap, nil
}
