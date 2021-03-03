test:
	go list ./... | grep -v extern | xargs go test -count 1 -cover -race -timeout 1m
