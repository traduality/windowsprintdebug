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

## License

Windows Print Debug by Traduality Language Solutions, Inc. is marked with CC0
1.0 Universal. To view a copy of this license, see [LICENSE.txt](./LICENSE.txt).
