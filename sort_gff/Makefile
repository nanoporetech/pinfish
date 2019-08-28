all: build

.PHONY: ct com push fetch gt fmt fix

TOOLS=../tools

# Utility targets:
ct:
	git log --graph
fmt: *.go
	go fmt *.go
com: fmt
	git commit -a
push:
	git push --all
fetch:
	git fetch --all
fix:
	go fix .

# Target:
BINARY=sort_gff

# These are the values we want to pass for VERSION and BUILD
VERSION=0.1.0
BUILD=`git rev-parse HEAD`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

# Builds the project
build: *.go
	go build ${LDFLAGS} -o ${BINARY}

# Installs our project: copies binaries
install:
	go install ${LDFLAGS}

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
# Run it on a small test gff:
test:
	./sort_gff test_data/small_test.gff
