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
BINARY=polish_clusters

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

# Test tool on simulated dataset (with sequencing errors):
test_sim:
	 ./polish_clusters -t 40 -c 100 -o test_data/cons.fas -a test_data/cls_sirv_sim_mm2.tab ../spliced_bam2gff/test_data/sirv_simulated.bam
	 (cd test_data; minimap2 -ax splice --splice-flank=no SIRV_150601a.fasta cons.fas | samtools view -Sb -F 2304 | samtools sort - -o test_sorted.bam)
	 (cd test_data; ../../spliced_bam2gff/spliced_bam2gff -M test_sorted.bam > test_pol.gff;gffcompare -r SIRV_C_150601a.gtf test_pol.gff)
	 cat test_data/gffcmp.stats

# Test tool on SIRV E0 full length cDNA dataset (may take a while):
test_sirv:
	 ./polish_clusters -t 40 -c 50 -o test_data/sirv_cons.fas -a test_data/sirv_e0.tab test_data/sirv_e0_sorted.bam
	 (cd test_data; minimap2 -ax splice --splice-flank=no SIRV_150601a.fasta sirv_cons.fas | samtools view -Sb -F 2304 | samtools sort - -o test_sorted.bam)
	 (cd test_data; ../../spliced_bam2gff/spliced_bam2gff -M test_sorted.bam > test_pol.gff;gffcompare -r SIRV_C_150601a.gtf test_pol.gff)
	 cat test_data/gffcmp.stats

