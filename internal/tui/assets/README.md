# Bundled Windows Terminal

`WindowsTerminal_x64.zip` is the official **unpackaged** Windows Terminal build,
embedded in the Windows executable so a standalone launch gets a capable
terminal even on machines without `winget` and without network access (Windows
Sandbox is the motivating case).

- Release: [Windows Terminal v1.25.2733.0](https://github.com/microsoft/terminal/releases/tag/v1.25.2733.0)
- Asset: `Microsoft.WindowsTerminal_1.25.2733.0_x64.zip`
- SHA-256: `bf3ef2012f6c44d8340a4c58125acc9498d19b580f9890dc043cdf831852e796`
- License: MIT. The archive includes `NOTICE.html`, which is extracted next to
  `WindowsTerminal.exe`.

To update it, download the new x64 archive from the latest stable release, put
it here under the same name, update the hash above and the constant in
`windowsterminal_windows_test.go`, then run the Windows tests. The download
fallback in `windowsterminal_portable_windows.go` still covers architectures
without a bundled archive.
