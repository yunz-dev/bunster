package ryuko

import "embed"

//go:embed runtime
var RuntimeFS embed.FS

//go:embed stubs
var StubsFS embed.FS
