package scp_test

import (
	"io/ioutil"
	"strings"

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

	AfterEach(func() {
		mockSSHServer.Stop()
	})

	XDescribe("#Connect", func() {
		// TODO: test errors: invalid creds, double connect, failed dial (bad endpoint)

		Describe("with valid credentials", func() {
			It("should successfully dial an SSH connection", func() {
				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				Expect(session.Close()).To(Succeed())
			})
		})
	})

	Describe("#Send", func() {
		// TODO: test errors:
		// - failed to open session (shut down test server after connect)
		// - failed to open stdin (skip if too difficult)
		// - failed to copy contents (skip if too difficult)
		// - failed to write zero byte (skip if too difficult)
		// - failed to run command (use bad base path)
		// At least one of the difficult error cases should be solvable
		// by adding an InvalidStdin bool to the test server.

		It("should send the provided contents and metadata", func(done Done) {
			go func() {
				defer GinkgoRecover()
				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				contents := ioutil.NopCloser(strings.NewReader("some-contents"))
				Expect(session.Send("/tmp/watch", contents, 0644, 100)).To(Succeed())
				Expect(session.Close()).To(Succeed())
				close(done)
			}()
			var result string
			Eventually(mockSSHServer.CommandChan).Should(Receive(&result))
			Expect(result).To(Equal("/usr/bin/scp -tr /tmp"))
			Eventually(mockSSHServer.DataChan).Should(Receive(&result))
			Expect(result).To(Equal("C0644 100 watch\nsome-contents\x00"))
		})
	})
})
