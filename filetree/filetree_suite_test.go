package filetree_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFileTree(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileTree Suite")
}
