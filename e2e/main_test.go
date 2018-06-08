package e2e

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/dlespiau/balance/e2e/harness"
	"github.com/dlespiau/balance/e2e/harness/logger"
)

var kube *harness.Harness

func manifestDirectory() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "manifests")
}

func TestMain(m *testing.M) {
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	noCleanup := flag.Bool("no-cleanup", false, "should test cleanup after themselves")
	verbose := flag.Bool("log.verbose", false, "turn on more verbose logging")
	interactive := flag.Bool("log.interactive", false, "print log messages as they happen")
	//tag := flag.String("image-tag", "", "tag of docker images to test")
	flag.Parse()

	options := harness.Options{
		Kubeconfig:        *kubeconfig,
		ManifestDirectory: manifestDirectory(),
		NoCleanup:         *noCleanup,
	}
	if *verbose {
		options.LogLevel = logger.Debug
	}
	if *interactive {
		options.Logger = &logger.PrintfLogger{}
	}
	kube = harness.New(options)

	if err := kube.Setup(); err != nil {
		log.Printf("failed to initialize test harness: %v\n", err)
	}

	code := m.Run()

	if err := kube.Close(); err != nil {
		log.Printf("failed to teardown test harness: %v\n", err)
		code = 1
	}

	os.Exit(code)
}
