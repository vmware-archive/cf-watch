package main

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
)

type Plugin struct{}

func (*Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("watch")
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
