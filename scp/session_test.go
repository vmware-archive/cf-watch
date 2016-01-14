package scp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/pivotal-cf/cf-watch/scp"
	"github.com/pivotal-cf/cf-watch/scp/mocks"
)

var _ = Describe("Session", func() {
	var (
		session       *Session
		mockSSHServer *mocks.SSHServer
		serverAddress string
	)

	BeforeEach(func() {
		session = &Session{}
		mockSSHServer = &mocks.SSHServer{
			User:     "some-valid-user",
			Password: "some-valid-password",
		}
		serverAddress = mockSSHServer.Start()
	})

	Describe("#Connect", func() {
		It("should dial an SSH session", func() {
			go func() {
				defer GinkgoRecover()
				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				Expect(session.Send()).To(Succeed())
			}()
			Expect(<-mockSSHServer.DataChan).To(Equal("test"))
			Expect(<-mockSSHServer.CommandChan).To(Equal("echo hi"))
		})
	})
})
