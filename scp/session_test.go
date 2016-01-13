package scp_test

import (
	"errors"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"

	. "github.com/pivotal-cf/cf-watch/scp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Session", func() {
	var (
		session   *Session
		serverURL string
	)

	BeforeEach(func() {
		session = &Session{}
		serverURL = "0.0.0.0:" + freePort()
		config := &ssh.ServerConfig{
			PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
				if c.User() == "some-valid-user" && string(pass) == "some-valid-password" {
					return nil, nil
				}
				return nil, errors.New("some error")
			},
		}
		_, err := net.Listen("tcp", serverURL)
		Expect(err).NotTo(HaveOccurred())
		Expect(config).NotTo(BeNil())
	})

	Describe("#Connect", func() {
		It("should dial an SSH session", func() {

		})
	})
})

func freePort() string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		Fail(err.Error())
	}
	defer listener.Close()

	address := listener.Addr().String()
	return strings.SplitN(address, ":", 2)[1]
}
