dev:
	go build -o logdna-gcp ./cmd
	@echo serving on localhost:8080
	env FUNCTION_TARGET=LogDNAUpload ./logdna-gcp
