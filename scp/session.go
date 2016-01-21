package scp

import (
	"errors"
	"fmt"
	"io"

	"github.com/pivotal-cf/cf-watch/filetree"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	client *ssh.Client
}

//go:generate mockgen -package mocks -destination mocks/file.go github.com/pivotal-cf/cf-watch/scp File
type File interface {
	filetree.File
}

func (s *Session) Connect(endpoint, username, password string) error {
	if s.client != nil {
		return errors.New("already connected")
	}

	var err error
	s.client, err = ssh.Dial("tcp", endpoint, &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) Close() error {
	if s.client == nil {
		return nil
	}
	if err := s.client.Close(); err != nil {
		return err
	}
	s.client = nil
	return nil
}

func (s *Session) Send(file File) error {
	defer file.Close()
	if s.client == nil {
		return errors.New("session closed")
	}

	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	errChan := make(chan error, 1)
	go func() {
		stdin, err := session.StdinPipe()
		if err != nil {
			errChan <- err
			return
		}
		defer stdin.Close()

		mode, err := file.Mode()
		if err != nil {
			panic(err)
		}
		size, err := file.Size()
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(stdin, "C%04o %d %s\n", mode, size, file.Basename())

		if _, err := io.Copy(stdin, file); err != nil {
			errChan <- err
			return
		}
		if _, err := stdin.Write([]byte{0}); err != nil {
			errChan <- err
			return
		}
	}()
	go func() {
		if err := session.Run("/usr/bin/scp -tr /home/vcap"); err != nil {
			errChan <- err
		}
		close(errChan)
	}()
	return <-errChan
}
