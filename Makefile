.PHONY: gen
gen: go-generate	

.PHONY: go-gen
go-gen:
	@mkdir -p pb
	@go generate ./...

