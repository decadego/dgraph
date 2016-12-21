// This script is used to load data into Dgraph from an RDF file by performing
// mutations using the HTTP interface.
//
// You can run the script like
// go build . && ./dgraphloader -r path-to-gzipped-rdf.gz
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/dgraph-io/dgraph/goclient/client"
	"github.com/dgraph-io/dgraph/rdf"
	"github.com/dgraph-io/dgraph/x"
)

var (
	files   = flag.String("r", "", "Location of rdf files to load")
	dgraph  = flag.String("d", "127.0.0.1:8080", "Dgraph server address")
	geoJSON = flag.String("json", "", "Json file from which to upload geo data")
)

// Reads a single line from a buffered reader. The line is read into the
// passed in buffer to minimize allocations. This is the preferred
// method for loading long lines which could be longer than the buffer
// size of bufio.Scanner.
func readLine(r *bufio.Reader, buf *bytes.Buffer) error {
	isPrefix := true
	var err error
	for isPrefix && err == nil {
		var line []byte
		// The returned line is an internal buffer in bufio and is only
		// valid until the next call to ReadLine. It needs to be copied
		// over to our own buffer.
		line, isPrefix, err = r.ReadLine()
		if err == nil {
			buf.Write(line)
		}
	}
	return err
}

// processFile sends mutations for a given gz file.
func processFile(file string, batch *client.BatchMutation) {
	fmt.Printf("\nProcessing %s\n", file)
	f, err := os.Open(file)
	x.Check(err)
	defer f.Close()
	gr, err := gzip.NewReader(f)
	x.Check(err)

	var buf bytes.Buffer
	bufReader := bufio.NewReader(gr)
	for {
		err = readLine(bufReader, &buf)
		if err != nil {
			break
		}
		nq, err := rdf.Parse(buf.String())
		if err != nil {
			log.Fatal("While parsing RDF: ", err)
		}
		buf.Reset()
		if err = batch.AddMutation(nq, client.SET); err != nil {
			log.Fatal("While adding mutation to batch: ", err)
		}
	}
	if err != io.EOF {
		x.Checkf(err, "Error while reading file")
	}
}

func printCounters(batch *client.BatchMutation) {
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		c := batch.Counter()
		rate := float64(c.Rdfs) / c.Elapsed.Seconds()
		fmt.Printf("[Request: %6d] Total RDFs done: %8d RDFs per second: %7.0f\r", c.Mutations, c.Rdfs, rate)
	}
}

func main() {
	x.Init()

	var err error
	conn, err := grpc.Dial(*dgraph, grpc.WithInsecure())
	if err != nil {
		log.Fatal("DialTCPConnection")
	}
	defer conn.Close()

	batch := client.NewBatchMutation(context.Background(), conn, 1000, 10)
	go printCounters(batch)

	if *geoJSON != "" {
		uploadJSON(*geoJSON, batch)
		return
	}

	filesList := strings.Split(*files, ",")
	x.AssertTrue(len(filesList) > 0)
	for _, file := range filesList {
		processFile(file, batch)
	}
	batch.Flush()

	c := batch.Counter()
	// Lets print an empty line, otherwise Number of Mutations overwrites the previous
	// printed line.
	fmt.Printf("%100s\r", "")
	fmt.Printf("Number of mutations run   : %d\n", c.Mutations)
	fmt.Printf("Number of RDFs processed  : %d\n", c.Rdfs)
	fmt.Printf("Time spent                : %v\n", c.Elapsed)
	fmt.Printf("RDFs processed per second : %d\n", c.Rdfs/uint64(c.Elapsed.Seconds()))
}
