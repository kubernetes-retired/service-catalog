package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func writeJSON(w io.Writer, obj interface{}, n int) {

	if n == 0 {
		//default the JSON indent to three spaces
		n = 3
	}
	indent := strings.Repeat(" ", n)
	j, err := json.MarshalIndent(obj, "", indent)
	if err != nil {
		fmt.Fprintf(w, "err marshaling json: %v\n", err)
		return
	}
	fmt.Fprint(w, string(j))
}
