package scp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSCP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SCP Suite")
}
