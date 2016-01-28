package scp_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-cf/cf-watch/filetree"
	. "github.com/pivotal-cf/cf-watch/scp"
	"github.com/pivotal-cf/cf-watch/scp/mocks"
)

var _ = Describe("Session", func() {
	var (
		session       *Session
		mockSSHServer *mocks.SSHServer
		serverAddress string
		mockCtrl      *gomock.Controller
		mockFileTree  *mocks.MockFile
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockFileTree = mocks.NewMockFile(mockCtrl)
		session = &Session{}
		mockSSHServer = &mocks.SSHServer{
			User:     "some-valid-user",
			Password: "some-valid-password",
		}
		serverAddress = mockSSHServer.Start()
	})

	AfterEach(func() {
		mockCtrl.Finish()
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
			mockFile := mocks.NewMockFile(mockCtrl)
			mockFile.EXPECT().Basename().Return("some-file")
			mockFile.EXPECT().Mode().Return(os.FileMode(0644), nil)
			mockFile.EXPECT().IsDir().Return(false, nil)
			mockFile.EXPECT().Close().Return(nil)
			mockFile.EXPECT().Read(gomock.Any()).Return(14, io.EOF).Do(func(buffer []byte) {
				defer GinkgoRecover()

				_, err := bytes.NewBufferString("some-contents").Read(buffer)
				Expect(err).NotTo(HaveOccurred())
			})

			mockFileTree.EXPECT().Basename().Return("some-dir")
			mockFileTree.EXPECT().Children().Return([]filetree.File{mockFile})
			mockFileTree.EXPECT().Mode().Return(os.FileMode(0755), nil)
			mockFileTree.EXPECT().Open().Return(nil)
			mockFileTree.EXPECT().IsDir().Return(true, nil)

			go func() {
				defer GinkgoRecover()

				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				defer session.Close()

				fmt.Println("before send")
				Expect(session.Send(mockFileTree)).To(Succeed())
				fmt.Println("after send")

				Expect(session.Close()).To(Succeed())
				close(done)
			}()

			Eventually(mockSSHServer.Data).Should(gbytes.Say("D0755 0 some-dir\n"))
			fmt.Println(string(mockSSHServer.Data.Contents()))
			//Eventually(mockSSHServer.Data).Should(gbytes.Say("C0644 14 some-file\nsome-contents\n"))
			//Eventually(mockSSHServer.Data).Should(gbytes.Say("E\n\x00"))

			var result string
			Eventually(mockSSHServer.CommandChan).Should(Receive(&result))
			Expect(result).To(Equal("/usr/bin/scp -tr /home/vcap"))
		})

		Context("when the session is not connected", func() {
			XIt("should return an error", func() {
				mockFileTree.EXPECT().Close().Return(nil)

				err := session.Send(mockFileTree)
				Expect(err).To(MatchError("session closed"))
			})
		})

		Context("when the SSH session cannot be established", func() {
			XIt("should return an error", func() {
				mockFileTree.EXPECT().Close().Return(nil)

				mockSSHServer.RejectSession = true

				Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
				defer session.Close()

				err := session.Send(mockFileTree)
				Expect(err).To(MatchError("ssh: rejected: connect failed (session rejected)"))
			})
		})

		Context("when the remote scp command fails", func() {
			XIt("should return an error", func(done Done) {
				mockFileTree.EXPECT().Basename().Return("some-file")
				mockFileTree.EXPECT().Mode().Return(os.FileMode(0644), nil)
				mockFileTree.EXPECT().Size().Return(int64(14), nil)
				mockFileTree.EXPECT().Close().Return(nil)
				mockFileTree.EXPECT().Read(gomock.Any()).Return(14, io.EOF).Do(func(buffer []byte) {
					defer GinkgoRecover()

					_, err := bytes.NewBufferString("some-contents").Read(buffer)
					Expect(err).NotTo(HaveOccurred())
				})

				mockSSHServer.CommandExitStatus = 1

				go func() {
					defer GinkgoRecover()

					Expect(session.Connect(serverAddress, "some-valid-user", "some-valid-password")).To(Succeed())
					defer session.Close()

					err := session.Send(mockFileTree)
					Expect(err).To(MatchError(ContainSubstring("Process exited with: 1")))

					close(done)
				}()

				Eventually(mockSSHServer.CommandChan).Should(Receive())
			})
		})
	})
})
