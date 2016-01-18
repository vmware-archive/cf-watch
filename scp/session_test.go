package scp_test

import (
	"io/ioutil"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	. "github.com/pivotal-cf/cf-watch/scp"
	"github.com/pivotal-cf/cf-watch/scp/mocks"
)

var _ = Describe("Session", func() {
	var (
		session       *Session
		mockSSHServer *mocks.SSHServer
		serverAddress string
		mockCtrl      *gomock.Controller
		mockFile      *mocks.MockFile
	)

	BeforeEach(func() {
		session = &Session{}
		mockSSHServer = &mocks.SSHServer{
			User:     "some-valid-user",
			Password: "some-valid-password",
		}
		serverAddress = mockSSHServer.Start()
		mockCtrl = gomock.NewController(GinkgoT())
		mockFile = mocks.NewMockFile(mockCtrl)
	})

	AfterEach(func() {
		mockSSHServer.Stop()
	})

	Describe("#Connect", func() {
		Context("with valid credentials", func() {
			It("should successfully dial an SSH connection", func() {
				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				Expect(session.Close()).To(Succeed())
			})
		})

		Context("with invalid credentials", func() {
			It("should return an error", func() {
				err := session.Connect(serverAddress, "some-invalid-user", "some-invalid-password")
				Expect(err).To(MatchError(ContainSubstring("ssh: unable to authenticate")))
			})
		})

		Context("when already connected", func() {
			It("should return an error", func() {
				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				defer session.Close()
				err := session.Connect(serverAddress, "some-valid-user", "some-valid-password")
				Expect(err).To(MatchError("already connected"))
			})
		})
	})

	Describe("#Close", func() {
		It("should allow a session to be re-connected", func() {
			Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
			Expect(session.Close()).To(Succeed())
			Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
			Expect(session.Close()).To(Succeed())
		})

		Context("when called on a closed session", func() {
			It("should succeed", func() {
				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				Expect(session.Close()).To(Succeed())
				Expect(session.Close()).To(Succeed())
			})
		})
	})

	Describe("#Send", func() {
		It("should create an empty directory", func(done Done) {
			mockFile.EXPECT().BaseName().Return("some-file")
			mockFile.EXPECT().Children().Return([]*File{})
			mockFile.EXPECT().ModePerm().Return("0644")
			mockFile.EXPECT().Read(gomock.Any()).Return(12, nil)
			mockFile.EXPECT().Close().Return(nil)
			go func() {
				defer GinkgoRecover()

				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				defer session.Close()

				contents := ioutil.NopCloser(strings.NewReader("some-contents"))
				Expect(session.Send("/tmp/watch", contents, 0644, 100)).To(Succeed())

				Expect(session.Close()).To(Succeed())
				close(done)
			}()

			Eventually(mockSSHServer.Data).Should(gbytes.Say("C0644 100 watch\nsome-contents\x00"))

			var result string
			Eventually(mockSSHServer.CommandChan).Should(Receive(&result))
			Expect(result).To(Equal("/usr/bin/scp -tr /tmp"))
		})

		Context("when the session is not connected", func() {
			It("should return an error", func() {
				contents := ioutil.NopCloser(strings.NewReader(""))
				err := session.Send("/tmp/watch", contents, 0644, 100)
				Expect(err).To(MatchError("session closed"))
			})
		})

		Context("when the SSH session cannot be established", func() {
			It("should return an error", func() {
				mockSSHServer.RejectSession = true

				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				defer session.Close()

				contents := ioutil.NopCloser(strings.NewReader(""))
				err := session.Send("/tmp/watch", contents, 0644, 100)
				Expect(err).To(MatchError("ssh: rejected: connect failed (session rejected)"))
			})
		})

		Context("when the remote scp command fails", func() {
			It("should return an error", func(done Done) {
				mockSSHServer.CommandExitStatus = 1

				go func() {
					defer GinkgoRecover()

					Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
					defer session.Close()

					contents := ioutil.NopCloser(strings.NewReader(""))
					err := session.Send("/tmp/watch", contents, 0644, 100)
					Expect(err).To(MatchError(ContainSubstring("Process exited with: 1")))

					close(done)
				}()

				Eventually(mockSSHServer.CommandChan).Should(Receive())
			})
		})
	})
})
