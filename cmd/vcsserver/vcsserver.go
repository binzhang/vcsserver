package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/sourcegraph/go-vcs"
	"github.com/sourcegraph/vcsserver"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var bindAddr = flag.String("http", ":8080", "HTTP bind address")
var storageRoot = flag.String("storage", "/tmp/vcsserver", "storage root dir for VCS repos")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "vcsserver mirrors and serves VCS repositories.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		fmt.Fprintf(os.Stderr, "\tvcsserver [options] request-path,host,repo-route,vcs-type,vcs-scheme\n\n")
		fmt.Fprintf(os.Stderr, "The options are:\n\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "Example usage:\n\n")
		fmt.Fprintf(os.Stderr, "\tTo run a mirror mapping http://localhost:8080/github.com/user/repo.git to\n")
		fmt.Fprintf(os.Stderr, "\tgit repos at git://github.com/user/repo.git:\n")
		fmt.Fprintf(os.Stderr, "\t    $ vcsserver '/github.com/,github.com,^/([^/]+)/([^/])+,git,git'\n\n")
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
	}

	log.SetPrefix("")
	log.SetFlags(0)

	fields := strings.Split(flag.Arg(0), ",")
	if want := 5; len(fields) != want {
		log.Fatalf("mapping must have %d comma-separated fields, got %d", want, len(fields))
	}

	var vcstype vcs.VCS
	var present bool
	if vcstype, present = vcs.VCSByName[fields[3]]; !present {
		log.Fatalf("unrecognized VCS type: %q", fields[1])
	}

	repo, err := regexp.Compile(fields[2])
	if err != nil {
		log.Fatalf("bad repo route regexp: %s", err)
	}

	m := vcsserver.Mapping{
		Host:   fields[1],
		VCS:    vcstype,
		Repo:   repo,
		Scheme: fields[4],
	}
	http.Handle(fields[0], handlers.CombinedLoggingHandler(os.Stderr, m))

	fmt.Fprintf(os.Stderr, "starting server on %s with mappings: %+v\n", *bindAddr, m)
	err = http.ListenAndServe(*bindAddr, nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %s", err)
	}
}
