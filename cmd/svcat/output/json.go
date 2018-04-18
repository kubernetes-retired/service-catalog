package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func writeJSON(w io.Writer, obj interface{}) {
	indent := strings.Repeat(" ", 3)
	j, err := json.MarshalIndent(obj, "", indent)
	if err != nil {
		fmt.Fprintf(w, "err marshaling json: %v\n", err)
		return
	}
	fmt.Fprint(w, string(j))
}
