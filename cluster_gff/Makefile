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
BINARY=cluster_gff

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

# Test tool on GFF generated from simulated data (with sequencing errors) aligned using minimap2:
test_sim:
	 ./cluster_gff -a ./test_data/cls_sirv_sim_mm2.tab -c 100  ./test_data/sirv_simulated_mm2.gff > ./test_data/cls_sirv_sim_mm2.gff
	 (cd test_data; gffcompare -r SIRV_C_150601a.gtf cls_sirv_sim_mm2.gff)
	 cat ./test_data/gffcmp.stats

# Test tool on GFF generated from SIRV E0 mix full lenegth cDNA dataset:
test_sirv:
	 ./cluster_gff -a ./test_data/sirv_e0.tab -c 10 ./test_data/sirv_e0_sorted.gff > ./test_data/cls_sirv_e0_sorted.gff
	 (cd test_data; gffcompare -r SIRV_C_150601a.gtf cls_sirv_e0_sorted.gff)
	 cat ./test_data/gffcmp.stats

# Run tool on small example GFF:
test_small:
	 ./cluster_gff -a test_data/rs_clusters.tab -c 1 test_data/real_small.gff
