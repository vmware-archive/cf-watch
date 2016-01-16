package watch

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

//go:generate mockgen -package mocks -destination mocks/session.go github.com/pivotal-cf/cf-watch/watch Session
type Session interface {
	Connect(endpoint, guid, password string) error
	Send(path string, contents io.ReadCloser, mode os.FileMode, size int64) error
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

	contents := strings.NewReader("test")
	if err := p.Session.Send("/tmp/watch", ioutil.NopCloser(contents), 0644, contents.Size()); err != nil {
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
