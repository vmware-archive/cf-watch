package watch

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/pivotal-cf/cf-watch/filetree"
)

//go:generate mockgen -package mocks -destination mocks/session.go github.com/pivotal-cf/cf-watch/watch Session
type Session interface {
	Connect(endpoint, guid, password string) error
	Send(file filetree.File) error
}

//go:generate mockgen -package mocks -destination mocks/cli.go github.com/pivotal-cf/cf-watch/watch CLI
type CLI interface {
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
}

//go:generate mockgen -package mocks -destination mocks/ui.go github.com/pivotal-cf/cf-watch/watch UI
type UI interface {
	Failed(message string, args ...interface{})
}

//go:generate mockgen -package mocks -destination mocks/file.go github.com/pivotal-cf/cf-watch/watch File
type File interface {
	filetree.File
}

//go:generate mockgen -package mocks -destination mocks/tree.go github.com/pivotal-cf/cf-watch/watch Tree
type Tree interface {
	New(path string) (filetree.File, error)
}

type Plugin struct {
	Session Session
	UI      UI
	Tree    Tree
}

func (p *Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	var cli CLI = cliConnection

	if len(args) < 3 {
		p.UI.Failed("Usage: cf %s <app> <local-dir>", args[0])
		return
	}

	appGUIDOutput, err := cli.CliCommandWithoutTerminalOutput("app", args[1], "--guid")
	if err != nil {
		p.UI.Failed("Failed to retrieve app GUID: %s", err)
		return
	}
	appGUID := strings.TrimSpace(appGUIDOutput[0])

	appJSONOutput, err := cli.CliCommandWithoutTerminalOutput("curl", path.Join("/v2/apps", appGUID))
	if err != nil {
		p.UI.Failed("Failed to retrieve app info: %s", err)
		return
	}

	var appInfo struct{ Entity struct{ Instances int } }
	if err := json.Unmarshal([]byte(appJSONOutput[0]), &appInfo); err != nil {
		p.UI.Failed("Failed to parse app info JSON: %s", err)
		return
	}

	if appInfo.Entity.Instances != 1 {
		p.UI.Failed("App must have exactly one instance to be used with cf-watch.")
		return
	}

	infoJSONOutput, err := cli.CliCommandWithoutTerminalOutput("curl", "/v2/info")
	if err != nil {
		p.UI.Failed("Failed to retrieve CC info: %s", err)
		return
	}

	var info struct {
		AppSSHEndpoint string `json:"app_ssh_endpoint"`
	}
	if err := json.Unmarshal([]byte(infoJSONOutput[0]), &info); err != nil {
		p.UI.Failed("Failed to parse CC info JSON: %s", err)
		return
	}

	passwordOutput, err := cli.CliCommandWithoutTerminalOutput("ssh-code")
	if err != nil {
		p.UI.Failed("Failed to retrieve SSH code: %s", err)
		return
	}

	username := fmt.Sprintf("cf:%s/0", appGUID)
	password := strings.TrimSpace(passwordOutput[0])
	if err := p.Session.Connect(info.AppSSHEndpoint, username, password); err != nil {
		p.UI.Failed("Failed to connect to app over SSH: %s", err)
		return
	}

	file, err := p.Tree.New(args[2])
	if err != nil {
		p.UI.Failed("Failed to process local app directory: %s", err)
		return
	}

	if err := p.Session.Send(file); err != nil {
		p.UI.Failed("Failed to send data to app over SSH: %s", err)
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
