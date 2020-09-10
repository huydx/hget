COMMIT = $$(git describe --always)

deps:
	@echo "====> Install dependencies..."
	go get -d github.com/fatih/color
	go get -d github.com/mattn/go-colorable
	go get -d github.com/mattn/go-isatty
	go get -d github.com/fatih/color
	go get -d gopkg.in/cheggaaa/pb.v1
	go get -d github.com/mattn/go-isatty
	go get -d github.com/imkira/go-task
	go get -d github.com/fujiwara/shapeio
	go get -d github.com/alecthomas/units

clean:
	@echo "====> Remove installed binary"
	rm -f bin/hget

install: deps
	@echo "====> Build hget in ./bin "
	go build -ldflags "-X main.GitCommit=\"$(COMMIT)\"" -o bin/hget
