# Development notes

xmasbot started as a fork of [ugjka/newyearsbot](https://github.com/ugjka/newyearsbot),
retargeted from New Year's Eve to Christmas. This document explains the changes
that turned it into an independent project rather than a derivative — what
changed, and why.

## Removing the ugjka dependencies

The fork originally depended on two packages maintained solely by ugjka
(the original author), which meant relying on a single person continuing to
maintain both an IRC client and a timezone-lookup library indefinitely.

### IRC: `github.com/ugjka/kittybot` → `github.com/lrstanley/girc`

`girc` (v1.1.1) is a stable, actively maintained, widely-used IRC client
library that covers everything kittybot provided:

- SASL PLAIN authentication (`girc.SASLPlain`)
- TLS with configurable verification skipping (`Config.TLSConfig`)
- Custom local-address binding via `Client.DialerConnect(dialer)` — any
  `*net.Dialer` already satisfies girc's `Dialer` interface
  (`Dial(network, address string) (net.Conn, error)`), so no custom dial
  function types were needed anymore
- Flood control (`Config.AllowFlood`)
- Safe message-length calculation (`Client.MaxEventLength()`)

One deliberate non-change: kittybot's `MsgMaxSize` computed the maximum safe
message length using the bot's *current* hostmask. girc's `MaxEventLength()`
instead uses the server's declared `NICKLEN`/`USERLEN`/`HOSTLEN` (from the
IRCv3 ISUPPORT/005 reply) to compute a worst-case `:nick!user@host ` prefix.
That's arguably more robust — it can't be invalidated by a cloak/vhost change
mid-connection — so no attempt was made to reproduce kittybot's exact
approach.

As a result of this swap, `main.go`'s custom dial/TLS-dial construction
(previously ~80 lines building `nyb.DialFunc`/`nyb.TLSDialFunc` closures) was
also simplified: dial construction now happens once inside `xmas.New()`,
since girc only needs a plain `*net.Dialer` plus a `*tls.Config`.

### Timezone lookup: `github.com/ugjka/go-tz` → `github.com/ringsaturn/tzf`

`tzf` resolves a lat/lon coordinate to an IANA timezone name
(`tzf.NewDefaultFinder()` + `GetTimezoneName(lng, lat)`), offline, with its
data embedded as a normal Go module dependency (no runtime network calls).

**Pinned to `v1.2.3`, not `@latest` (`v1.2.4`)** — the `v1.2.4` release's
`go.mod` requires `github.com/paulmach/orb@v0.14.0`, a version that was never
actually published to the module proxy. `go get` fails to resolve it. `v1.2.3`
resolves cleanly.

This swap touched `xmas/actions.go` (the `!xmas <location>`/`!time <location>`
handlers) and the dev-only data-generation tools (`utils/tzbuilder`,
`utils/validatetz`, `utils/validatetz/interactive`). The `interactive` tool's
`-ext` flag (load a custom GeoJSON boundary file) now uses
`tzf.NewFinderFromRawJSON` with the GeoJSON unmarshaled into
`tzf/convert.BoundaryFile`, replacing go-tz's `LoadGeoJSON`.

## Package rename: `nyb` → `xmas`

The package was named `nyb` ("new year bot") from the original project. Since
this is a Christmas bot with no New Year's code path left, it was renamed to
`xmas` (folder, package declaration, import path, and the
`utils/tzbuilder`-generated source template that writes the zone-data file).

## Release automation

`.github/workflows/release.yml` tags every push to `main` with a
`YYYYMMDD-HHMMSS` UTC timestamp, cross-compiles release binaries via
`makerel` (~40 OS/arch combinations), and publishes them as a GitHub Release.

`makerel/main.go` previously `panic()`'d on the first target that failed to
build, which would abort the entire release job over a single unsupported
platform (this bit hnybot in practice: `linux/loong64` needs Go 1.19+ and
`freebsd/riscv64` needs Go 1.21+ — an older pinned Go toolchain took the whole
release down). It now logs and skips a failed target instead. `go.mod`/
`go.work` are on Go 1.25, well past what any current build target needs.

## Everything else

`badoux/checkmail` and `fatih/color` (used only for email-format validation
and CLI color output) are not ugjka packages and were left as-is — they're
small, standard, low-risk dependencies, not part of the "no single-author
lock-in" concern this cleanup addressed.
