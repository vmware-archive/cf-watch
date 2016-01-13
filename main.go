package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	"golang.org/x/crypto/ssh"
)

type Plugin struct{}

func cliPanic(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func (*Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	appGUIDOutput, err := cliConnection.CliCommandWithoutTerminalOutput("app", args[1], "--guid")
	if err != nil {
		cliPanic(err)
	}
	appGUID := strings.TrimSpace(appGUIDOutput[0])

	sshPasswordOutput, err := cliConnection.CliCommandWithoutTerminalOutput("ssh-code")
	if err != nil {
		cliPanic(err)
	}
	sshPassword := strings.TrimSpace(sshPasswordOutput[0])

	infoJSONOutput, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v2/info")
	if err != nil {
		cliPanic(err)
	}

	var info struct {
		AppSSHEndpoint string `json:"app_ssh_endpoint"`
	}
	err = json.Unmarshal([]byte(infoJSONOutput[0]), &info)
	if err != nil {
		cliPanic(err)
	}

	clientConfig := &ssh.ClientConfig{
		User: fmt.Sprintf("cf:%s/0", appGUID),
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
	}
	client, err := ssh.Dial("tcp", info.AppSSHEndpoint, clientConfig)
	if err != nil {
		cliPanic(err)
	}
	session, err := client.NewSession()
	if err != nil {
		cliPanic(err)
	}

	defer session.Close()
	sessionIn, err := session.StdinPipe()
	if err != nil {
		cliPanic(err)
	}

	go func() {
		defer sessionIn.Close()
		content := "watch file\n"
		fmt.Fprintln(sessionIn, "C0644", len(content), "watch")
		fmt.Fprint(sessionIn, content)
		fmt.Fprint(sessionIn, "\x00")
	}()
	if err := session.Run("/usr/bin/scp -tr /tmp"); err != nil {
		cliPanic(err)
	}
}

func (*Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Watch",
		Commands: []plugin.Command{
			plugin.Command{
				Name: "watch",
			},
		},
	}
}

func main() {
	plugin.Start(&Plugin{})
}
