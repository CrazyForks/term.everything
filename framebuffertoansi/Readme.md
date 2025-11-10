Convert a framebuffer (bytes) with pixel data into ANSI escape codes for terminal display.

## Usage

```go
drawState := framebuffertoansi.MakeDrawState(true)
var grid []byte = ...
width := 80
height := 24
drawState.DrawDesktop(grid, uint32(width), uint32(height), "Hello, world!")
```