# Windows Print Debug

This Go package redirects various Go output to OutputDebugString for viewing in
a Windows debugger.

This Go package is intended for GUI Windows programs, not for console
applications.

## Features

| Feature      | Visible in debugger? | Comments                                                                    |
|--------------|----------------------|-----------------------------------------------------------------------------|
| `panic`      | ✅ Yes                | Output is also written to the file given to `runtime/debug.SetCrashOutput`. |
| log          | ✅ Yes                |                                                                             |
| log/slog     | ✅ Yes                |                                                                             |
| `print`      | ✅ Yes                |                                                                             |
| `os.Stdout`  | ❌ No                 |                                                                             |
| `os.Stderr`  | ❌ No                 |                                                                             |
| `fmt.Printf` | ❌ No                 |                                                                             |

## Support

Operating systems: Windows 11

Architectures: ARM64, x64
