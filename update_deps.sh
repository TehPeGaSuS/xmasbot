#!/bin/bash
# Update all the dependencies at once
#
# NOTE: github.com/ringsaturn/tzf is pinned to v1.2.3. Newer releases (e.g.
# v1.2.4) ship a go.mod that requires github.com/paulmach/orb@v0.14.0, a
# version that was never published to the module proxy, so `go get`/`go mod
# tidy` fails to resolve it. This script re-pins tzf back to v1.2.3 after
# updating everything else, so it doesn't silently reintroduce that break.
# Remove the re-pin once tzf ships a release with a valid go.mod.
find . -name go.mod -print -execdir sh -c '
	pwd
	go get -u ./... && go mod tidy
	if grep -q "ringsaturn/tzf v" go.mod 2>/dev/null; then
		echo "re-pinning github.com/ringsaturn/tzf to v1.2.3 (see script comment)"
		go get github.com/ringsaturn/tzf@v1.2.3 && go mod tidy
	fi
' \;
