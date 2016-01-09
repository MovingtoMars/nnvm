all:
	@go install github.com/MovingtoMars/nnvm/cmd/nnvmex

generate:
	go generate github.com/MovingtoMars/nnvm/...

wc:
	wc {cmd/nnvmex,types,ssa{,/analysis,/validate},target{/platform,/amd64}}/*.go
