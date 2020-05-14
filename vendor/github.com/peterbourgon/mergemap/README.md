# mergemap

mergemap is a Go library to recursively merge JSON maps.

[![Build Status](https://drone.io/github.com/peterbourgon/mergemap/status.png)](https://drone.io/github.com/peterbourgon/mergemap/latest)


## Behavior

mergemap performs a simple merge of the **src** map into the **dst** map. That
is, it takes the **src** value when there is a key conflict.

The only special behavior is when the conflicting key represents a map in both
src and dst. Then, mergemap recursively descends into both maps, repeating the
same logic. The max recursion depth is set by **mergemap.MaxDepth**.


## Usage

```go
var m1, m2 map[string]interface{}
json.Unmarshal(buf1, &m1)
json.Unmarshal(buf2, &m2)

merged := mergemap.Merge(m1, m2)
```

See the test file for some pretty straightforward examples.

