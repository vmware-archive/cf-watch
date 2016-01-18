package scp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	client *ssh.Client
}

//go:generate mockgen -package mocks -destination mocks/file.go github.com/pivotal-cf/cf-watch/scp File
type File interface {
	BaseName() string
	Children() ([]*File, error)
	ModePerm() (string, error)
	Read([]byte) (int, error)
	Close() error
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

func (s *Session) Send(path string, contents io.ReadCloser, mode os.FileMode, size int64) error {
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

		fmt.Fprintf(stdin, "C%04o %d %s\n", mode, size, filepath.Base(path))

		if _, err := io.Copy(stdin, contents); err != nil {
			errChan <- err
			return
		}
		if _, err := stdin.Write([]byte{0}); err != nil {
			errChan <- err
			return
		}
	}()
	go func() {
		if err := session.Run(fmt.Sprintf("/usr/bin/scp -tr %s", filepath.Dir(path))); err != nil {
			errChan <- err
		}
		close(errChan)
	}()
	return <-errChan
}
