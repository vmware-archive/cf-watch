package scp

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	client *ssh.Client
}

func (s *Session) Connect(endpoint, username, password string) error {
	if s.client != nil {
		panic("double connect")
	}

	var err error
	s.client, err = ssh.Dial("tcp", endpoint, &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	})
	if err != nil {
		panic(err)
	}

	return nil
}

func (s *Session) Send() error {
	session, err := s.client.NewSession()
	defer session.Close()
	go func() {
		stdin, err := session.StdinPipe()
		if err != nil {
			panic(err)
		}
		defer stdin.Close()
		fmt.Fprintln(stdin, "test")
	}()
	session.Run("echo hi")
	return err
}
