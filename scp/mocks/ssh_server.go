package mocks

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

type SSHServer struct {
	User        string
	Password    string
	CommandChan chan string
	DataChan    chan string
	listener    net.Listener
}

func (s *SSHServer) Start() (address string) {
	Expect(s.listener).To(BeNil(), "test server already started")

	var err error
	s.listener, err = net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())
	s.CommandChan = make(chan string)
	s.DataChan = make(chan string)
	go s.listen()
	return s.listener.Addr().String()
}

func (s *SSHServer) Stop() {
	Expect(s.listener.Close()).To(Succeed())
	s.listener = nil
}

func (s *SSHServer) listen() {
	defer GinkgoRecover()

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == s.User && string(pass) == s.Password {
				return nil, nil
			}
			return nil, errors.New("failed to start test ssh server")
		},
	}

	privateKey, err := ssh.ParsePrivateKey([]byte(sshPrivateKey))
	Expect(err).NotTo(HaveOccurred())
	config.AddHostKey(privateKey)

	for {
		tcpConn, err := s.listener.Accept()
		Expect(err).NotTo(HaveOccurred())
		_, newChannels, requests, err := ssh.NewServerConn(tcpConn, config)
		Expect(err).NotTo(HaveOccurred())

		go ssh.DiscardRequests(requests)
		go func() {
			for newChannel := range newChannels {
				go s.handleChannel(newChannel)
			}
		}()
	}
}

func (s *SSHServer) handleChannel(newChannel ssh.NewChannel) {
	defer GinkgoRecover()

	Expect(newChannel.ChannelType()).To(Equal("session"))

	channel, requests, err := newChannel.Accept()
	defer channel.Close()
	Expect(err).NotTo(HaveOccurred())
	go func() {
		defer GinkgoRecover()

		for request := range requests {
			Expect(request.Type).To(Equal("exec"))
			payloadLen := binary.BigEndian.Uint32(request.Payload[:4])
			Expect(request.Payload).To(HaveLen(int(payloadLen) + 4))
			Expect(request.Reply(true, nil)).To(Succeed())
			channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			s.CommandChan <- string(request.Payload[4:])
			// move this up ^
		}
	}()
	data, err := ioutil.ReadAll(channel)
	Expect(err).NotTo(HaveOccurred())
	s.DataChan <- string(data)
}
