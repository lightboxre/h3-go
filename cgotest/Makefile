# Makefile for native-vs-CGO benchmark comparison.
#
# Usage:
#   make bench               # run all paired benchmarks and compare
#   make bench BENCH=GridDisk  # filter to BenchmarkGridDisk_Native / _CGO
#   make bench COUNT=20      # more iterations for tighter confidence intervals
#   make bench-native        # one-shot native run (no comparison)
#   make bench-cgo           # one-shot CGO run (no comparison)

COUNT ?= 10
BENCH ?=

# Hidden temp files; cleaned up automatically by `bench` and on request by `clean`.
_N_RAW := .bench_native_raw.txt
_C_RAW := .bench_cgo_raw.txt
_N     := .bench_native.txt
_C     := .bench_cgo.txt

.PHONY: bench bench-native bench-cgo clean

bench:
	CGO_ENABLED=1 go test -bench='$(BENCH)_Native' -benchmem -count=$(COUNT) ./... > $(_N_RAW)
	CGO_ENABLED=1 go test -bench='$(BENCH)_CGO'    -benchmem -count=$(COUNT) ./... > $(_C_RAW)
	sed 's/_Native//' $(_N_RAW) > $(_N)
	sed 's/_CGO//'    $(_C_RAW) > $(_C)
	benchstat $(_C) $(_N) > benchmarks/benchmark.txt
	$(RM) $(_N_RAW) $(_C_RAW) $(_N) $(_C)

bench-native:
	CGO_ENABLED=1 go test -bench='$(BENCH)_Native' -benchmem -count=$(COUNT) ./...

bench-cgo:
	CGO_ENABLED=1 go test -bench='$(BENCH)_CGO' -benchmem -count=$(COUNT) ./...

clean:
	$(RM) $(_N_RAW) $(_C_RAW) $(_N) $(_C)
