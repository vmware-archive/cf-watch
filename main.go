package main

import (
	"os"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/pivotal-cf/cf-watch/scp"
	"github.com/pivotal-cf/cf-watch/watch"
)

func main() {
	plugin.Start(&watch.Plugin{
		Session: &scp.Session{},
		UI:      terminal.NewUI(os.Stdin, terminal.NewTeePrinter()),
	})
}
