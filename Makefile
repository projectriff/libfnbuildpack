.PHONY: clean build
GO_SOURCES = $(shell find . -type f -name '*.go')


build: scratch/io/projectriff/riff/io.projectriff.riff

scratch/io/projectriff/riff/io.projectriff.riff: bin/package buildpack.toml
	rm -fR $@ 							&& \
	./bin/package scratch 				&& \
	mkdir $@/latest 					&& \
	tar -C $@/latest -xzf $@/*/*.tgz


bin/package: go.mod $(GO_SOURCES)
	go build -i -ldflags='-s -w' -o bin/package package/main.go

clean:
	rm -fR scratch/
