# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	example.com/temporary-share-gateway/cmd/share-gateway	[no test files]
?   	example.com/temporary-share-gateway/internal/clock	[no test files]
ok  	example.com/temporary-share-gateway/internal/config	0.003s
ok  	example.com/temporary-share-gateway/internal/gateway	0.020s
ok  	example.com/temporary-share-gateway/internal/metrics	0.001s
ok  	example.com/temporary-share-gateway/internal/model	0.001s
ok  	example.com/temporary-share-gateway/internal/persist	0.014s
ok  	example.com/temporary-share-gateway/internal/security	0.002s
--- FAIL: TestShareCountAllowsOne (0.00s)
    service_test.go:53: allowed calls = 2
FAIL
FAIL	example.com/temporary-share-gateway/internal/share	0.018s
ok  	example.com/temporary-share-gateway/internal/transport	0.001s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/share-gateway): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/share-gateway): exit `0`
