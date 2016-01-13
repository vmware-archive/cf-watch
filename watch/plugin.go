package watch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

//go:generate mockgen -package mocks -destination mocks/session.go github.com/pivotal-cf/cf-watch/watch Session
type Session interface {
	Connect(endpoint, guid, password string) error
}

//go:generate mockgen -package mocks -destination mocks/cli.go github.com/pivotal-cf/cf-watch/watch CLI
type CLI interface {
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
}

//go:generate mockgen -package mocks -destination mocks/ui.go github.com/pivotal-cf/cf-watch/watch UI
type UI interface {
	Failed(message string, args ...interface{})
}

type Plugin struct {
	Session Session
	UI      UI
}

func (p *Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	var cli CLI = cliConnection
	appGUIDOutput, err := cli.CliCommandWithoutTerminalOutput("app", args[1], "--guid")
	if err != nil {
		p.UI.Failed("Failed to retrieve app GUID:", err)
		return
	}
	appGUID := strings.TrimSpace(appGUIDOutput[0])
	username := fmt.Sprintf("cf:%s/0", appGUID)

	passwordOutput, err := cli.CliCommandWithoutTerminalOutput("ssh-code")
	if err != nil {
		p.UI.Failed("Failed to retrieve SSH code:", err)
		return
	}
	password := strings.TrimSpace(passwordOutput[0])

	infoJSONOutput, err := cli.CliCommandWithoutTerminalOutput("curl", "/v2/info")
	if err != nil {
		p.UI.Failed("Failed to retrieve CC info:", err)
		return
	}

	var info struct {
		AppSSHEndpoint string `json:"app_ssh_endpoint"`
	}

	if err := json.Unmarshal([]byte(infoJSONOutput[0]), &info); err != nil {
		p.UI.Failed("Failed to parse CC info JSON:", err)
		return
	}

	if err := p.Session.Connect(info.AppSSHEndpoint, username, password); err != nil {
		p.UI.Failed("Failed to connect to app over SSH:", err)
		return
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
