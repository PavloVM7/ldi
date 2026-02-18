revive:
	$(GOPATH)/bin/revive -config ./revive.toml -formatter friendly -exclude di_test.go ./...

test:
	go test ./... -v -race -cover

check: revive test