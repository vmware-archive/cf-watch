package mocks

import (
	"errors"
	"io/ioutil"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

const privateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA+0j/PxGHo4oa6m+arKZhv2dL690XwG6ys5ZQq1laelWFBekL
c5FakIblim/g6ggZETPGhs9Cr+acUFOGoO5vkJ5+63d9+Moms4w41UjeZRNveENE
oae3/F1aZk43LRHghCd7UF+JtcLxIbSIPrcRoirj/rinbs6zsKJZT4uHdSsbodfz
HbLkASF6SuZs1e4WMEcmnrLbXI+12x2rZnW4libfU/Z8GWRyifFWUSKjFQZsveON
464wgMs2Wbs69voSeqA7awyven7i2ymFsD0StoTcbjFjqv5INa/K7PBrsVDcgd3H
tVQ7oRy6GyCxkFV8u6LOVxiIiCg8+825KUqTXQIDAQABAoIBADTiwx2h8dsgeNO4
U2RczBu9gMQOTy5n3eJgE3BMqPcwQoPg7VEQWXArg+nj7AE1XRk6vWCoBFADCAj7
20zJgd99DBdAmdmfqg+FxnxVDsFVGtPDzJD9PIK3nwwDECfDKG6H5LMguFnxwlAm
r7oLS4HG5x83+70dccIOGR/drM+irTYhhVaq0fSZz9dvLfg/MRZr+cudRSXUUfrk
KKVZidn20vErPIo5ZKHS5C2XcuNdMzUT1s5p51kJm8TFuWiQ4nWjoLToKtvgbyid
5mCBW4p2RjYw3LWjGWuaeL0Jm1qnVUz+U+TTLE6/enplitghYr5AJZ3ycSUiv5r9
BnDdaCECgYEA/pi1SW+S/lUPKM62UHK2pExtJ8q3prbMqCYU1ptlEFFca5REEjJQ
Is23a+IaXycN9pUGf1kcp0cjp5ZScsPRhXH9UbJ/DCTKWbkPYV+TpAOxBYXrujy7
R8wq/kdwTgPW9d0hFEcqEiigX+tkQY1LS44Urje5GTLbDG+OJRg2UlUCgYEA/Kud
pvLGjj7/eLd3998yACRiTAu8FkgBh892KLZ1u2nXx7JLDsOJtP7Pb6sBzCoakIKG
1Bv5WOoNB5BvG5EMeA0SFPADaA2Jt37YlmpRH1a3ipRSv6NPZhJTQYwfUUnepW+I
NwTVrHz2dItPufFnRgcZwYBPsbZMfTmDnKS/FOkCgYBRHDeNSMWMz25/8rM0mAdF
+q8/4R53N3+mBlPXNzSQaUtHXrn9DhhnriBEd4ktTVTufPXP9oThahGa35Iuy+Hh
YLpyn6pIJSRuRz32KKvxsddgyhSahaSosAv2bK4DvMdsFuHmAvINTPIi/Ow40hnt
3TsLcec/dutAX/3qJXeQ7QKBgFoO1TdPKwRCYg5l3mXD8O9qCHswZ47NhXYhtOzX
8+ij1hxAaU5O1cNkWw1jN1XM4AEH9QSfH+XYLmK20VNTBM25YuuBjMVGpgJ4PLyI
EngIEY1cRo41qDQqbfBcAEGaAbiXo0Zw+7PqKnHpwbX13ChymXSFxmICJwsvN8Da
W50ZAoGAaOy4Uie4IbUz5jjw1L/sKrAKelwyV89Hniihtp0/0ev+YMoFwPi/UmxN
4JMepDg3cXC9e3JC+Q7YI0EER+PfbGBE6L5HmlDPvg4AmgWpWD3O9wqcqYgjGh+U
Rj2jsQ6SVAstGW/M4w0EA0KkD2gEX1wOJIhQ+4wUb04wlg7c8jk=
-----END RSA PRIVATE KEY-----
`

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

	privateHostKey, err := ssh.ParsePrivateKey([]byte(privateKey))
	Expect(err).NotTo(HaveOccurred())
	config.AddHostKey(privateHostKey)

	for {
		tcpConn, err := s.listener.Accept()
		Expect(err).NotTo(HaveOccurred())
		_, newChannels, requests, err := ssh.NewServerConn(tcpConn, config)
		Expect(err).NotTo(HaveOccurred())

		go ssh.DiscardRequests(requests)
		go func() {
			defer GinkgoRecover()

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
	Expect(err).NotTo(HaveOccurred())
	go func() {
		for request := range requests {
			s.CommandChan <- string(request.Payload)
			Expect(request.Reply(true, nil)).To(Succeed())
		}
	}()
	data, err := ioutil.ReadAll(channel)
	Expect(err).NotTo(HaveOccurred())
	s.DataChan <- string(data)
}
