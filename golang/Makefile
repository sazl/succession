all: main

wiki_api:
	$(MAKE) -C proto

main: main.go proto
	go build