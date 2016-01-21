package watch_test

import (
	"errors"

	cliplugin "github.com/cloudfoundry/cli/plugin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/pivotal-cf/cf-watch/watch"
	"github.com/pivotal-cf/cf-watch/watch/mocks"
)

type cliConnectionWrapper struct {
	cliplugin.CliConnection
}

type mockCLIWrapper struct {
	*mocks.MockCLI
	cliConnectionWrapper
}

var _ = Describe("Plugin", func() {
	var (
		plugin      *Plugin
		mockCtrl    *gomock.Controller
		mockSession *mocks.MockSession
		mockCLI     *mockCLIWrapper
		mockUI      *mocks.MockUI
		mockFile    *mocks.MockFile
		mockTree    *mocks.MockTree
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockSession = mocks.NewMockSession(mockCtrl)
		mockCLI = &mockCLIWrapper{MockCLI: mocks.NewMockCLI(mockCtrl)}
		mockUI = mocks.NewMockUI(mockCtrl)
		mockFile = mocks.NewMockFile(mockCtrl)
		mockTree = mocks.NewMockTree(mockCtrl)
		plugin = &Plugin{
			Session: mockSession,
			UI:      mockUI,
			Tree:    mockTree,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("#Run", func() {
		It("should connect to the app and send /tmp/watch file with contents from local file", func() {
			mockTree.EXPECT().New("some-path").Return(mockFile, nil)

			mockSession.EXPECT().Connect("some-endpoint", "cf:some-guid/0", "some-password").Return(nil)
			mockSession.EXPECT().Send(mockFile).Return(nil)

			mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
			mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{`{"entity": {"instances": 1}}` + "\n"}, nil)
			mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{`{"app_ssh_endpoint": "some-endpoint"}` + "\n"}, nil)
			mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return([]string{"some-password\n"}, nil)

			plugin.Run(mockCLI, []string{"watch", "some-app", "some-path"})
		})

		Context("when watch is called without enough arguments", func() {
			It("should output a failure message", func() {
				mockUI.EXPECT().Failed("Usage: cf %s <app> <local-dir>", "watch")

				plugin.Run(mockCLI, []string{"watch"})
			})
		})

		Context("when the app GUID is unavailable", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return(nil, errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to retrieve app GUID: %s", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-file"})
			})
		})

		Context("when the app info is unavailable", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return(nil, errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to retrieve app info: %s", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-file"})
			})
		})

		Context("when the app info is not valid JSON", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{"some invalid JSON"}, nil)

				mockUI.EXPECT().Failed("Failed to parse app info JSON: %s", gomock.Any()).Do(func(_ string, args ...interface{}) {
					Expect(args[0]).To(MatchError("invalid character 's' looking for beginning of value"))
				})

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-file"})
			})
		})

		Context("when there is not exactly one instance of the app", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{`{"entity": {"instances": 2}}` + "\n"}, nil)

				mockUI.EXPECT().Failed("App must have exactly one instance to be used with cf-watch.")

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-file"})
			})
		})

		Context("when the CC info is unavailabe", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{`{"entity": {"instances": 1}}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return(nil, errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to retrieve CC info: %s", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-file"})
			})
		})

		Context("when the CC info is not valid JSON", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{`{"entity": {"instances": 1}}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{"some invalid JSON"}, nil)

				mockUI.EXPECT().Failed("Failed to parse CC info JSON: %s", gomock.Any()).Do(func(_ string, args ...interface{}) {
					Expect(args[0]).To(MatchError("invalid character 's' looking for beginning of value"))
				})

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-file"})
			})
		})

		Context("when the SSH code is unavailabe", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{`{"entity": {"instances": 1}}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{`{"app_ssh_endpoint": "some-endpoint"}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return(nil, errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to retrieve SSH code: %s", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-file"})
			})
		})

		Context("when connecting to the app over SSH fails", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{`{"entity": {"instances": 1}}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{`{"app_ssh_endpoint": "some-endpoint"}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return([]string{"some-password\n"}, nil)

				mockSession.EXPECT().Connect("some-endpoint", "cf:some-guid/0", "some-password").Return(errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to connect to app over SSH: %s", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-file"})
			})
		})

		Context("when creating a new tree fails", func() {
			It("should output a failure message", func() {
				mockTree.EXPECT().New("some-bad-path").Return(nil, errors.New("some error"))

				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{`{"entity": {"instances": 1}}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{`{"app_ssh_endpoint": "some-endpoint"}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return([]string{"some-password\n"}, nil)

				mockSession.EXPECT().Connect("some-endpoint", "cf:some-guid/0", "some-password").Return(nil)

				mockUI.EXPECT().Failed("Failed to process local app directory: %s", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-bad-path"})
			})
		})

		Context("when creating new directory over SSH fails", func() {
			It("should output a failure message", func() {
				mockTree.EXPECT().New("some-path").Return(mockFile, nil)

				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/apps/some-guid").Return([]string{`{"entity": {"instances": 1}}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{`{"app_ssh_endpoint": "some-endpoint"}` + "\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return([]string{"some-password\n"}, nil)

				mockSession.EXPECT().Connect("some-endpoint", "cf:some-guid/0", "some-password").Return(nil)
				mockSession.EXPECT().Send(mockFile).Return(errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to send data to app over SSH: %s", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app", "some-path"})
			})
		})
	})

	Describe("#GetMetadata", func() {
		It("should return plugin metadata", func() {
			Expect(plugin.GetMetadata()).To(Equal(cliplugin.PluginMetadata{
				Name: "Watch",
				Commands: []cliplugin.Command{
					cliplugin.Command{
						Name: "watch",
					},
				},
			}))
		})
	})
})
