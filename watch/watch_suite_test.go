package watch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Watch Suite")
}
