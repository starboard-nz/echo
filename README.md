# I/O wrapper for debugging

## net.Conn wrapper

```
import (
        "net"
        "github.com/starboard-nz/echo"
)

...

        conn, err := net.Dial("tcp", "golang.org:80")

        // wrap conn in echo.Conn and use econn instead of conn
        econn := echo.Conn{Conn: conn}
        
        econn.AddFileWrite("/path/to/file.log")
        w := econn.AddFileWrite("/path/to/file.go")
        // format read/write buffers as Go slices
        w.Go = true

        // also write to stderr
        econn.AddConsoleWriter()
```
