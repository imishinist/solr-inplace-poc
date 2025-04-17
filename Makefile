
.PHONY: build
build:
	go build -trimpath -ldflags "-s -w"

.PHONY: clean
clean:
	rm solr-inplace-poc
