release: bin rel
	@echo -n "What's the name/version number of this release?: "; \
	read VERSION; \
	COMMIT="`git log --abbrev-commit --oneline -n 1 | sed s/\'//g | sed 's/\\\"//g' | cut -d' ' -f1`"; \
	DATE="`date -u +'%Y-%m-%dT%H:%M:%SZ'`"; \
	RUNTIME="`go version | sed -En 's/.*(go[^ ]+).*/\1/p'`"; \
	LDFLAGS="-X main.buildVersion='$$VERSION' -X main.buildCommit='$$COMMIT' -X main.buildDate='$$DATE' -X main.buildRuntime='$$RUNTIME'"; \
	mkdir -p bin/gmem-$$VERSION; \
	cp README.md bin/gmem-$$VERSION; \
	for platform in darwin freebsd linux windows; do \
		for arch in 386 amd64; do \
			echo $$platform $$arch; \
			GOOS=$$platform GOARCH=$$arch go build -ldflags "$$LDFLAGS" -o bin/gmem-$$VERSION/gmem-"$$platform"-"$$arch" gmem.go; \
		done; \
	done; \
	cd bin; \
	echo "Tar-ing into rel/gmem-$$VERSION.tar.gz"; \
	tar cvzf ../rel/gmem-$$VERSION.tar.gz gmem-$$VERSION

rel:
	mkdir rel

bin:
	mkdir bin

clean:
	rm -rf bin
	rm -rf rel
