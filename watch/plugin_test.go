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
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockSession = mocks.NewMockSession(mockCtrl)
		mockCLI = &mockCLIWrapper{MockCLI: mocks.NewMockCLI(mockCtrl)}
		mockUI = mocks.NewMockUI(mockCtrl)
		plugin = &Plugin{
			Session: mockSession,
			UI:      mockUI,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("#Run", func() {
		It("should connect to the app", func() {
			mockSession.EXPECT().Connect("some-endpoint", "cf:some-guid/0", "some-password").Return(nil)

			mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
			mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return([]string{"some-password\n"}, nil)
			mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{`{"app_ssh_endpoint": "some-endpoint"}` + "\n"}, nil)

			plugin.Run(mockCLI, []string{"watch", "some-app"})
		})

		Describe("when the app GUID is unavailabe", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return(nil, errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to retrieve app GUID:", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app"})
			})
		})

		Describe("when the SSH code is unavailabe", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return(nil, errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to retrieve SSH code:", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app"})
			})
		})

		Describe("when the CC info is unavailabe", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return([]string{"some-password\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return(nil, errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to retrieve CC info:", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app"})
			})
		})

		Describe("when the CC info is not valid JSON", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return([]string{"some-password\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{"some invalid JSON"}, nil)

				mockUI.EXPECT().Failed("Failed to parse CC info JSON:", gomock.Any()).Do(func(_ string, args ...interface{}) {
					Expect(args[0]).To(MatchError("invalid character 's' looking for beginning of value"))
				})

				plugin.Run(mockCLI, []string{"watch", "some-app"})
			})
		})

		Describe("when the connecting to the app over SSH fails", func() {
			It("should output a failure message", func() {
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("app", "some-app", "--guid").Return([]string{"some-guid\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("ssh-code").Return([]string{"some-password\n"}, nil)
				mockCLI.EXPECT().CliCommandWithoutTerminalOutput("curl", "/v2/info").Return([]string{`{"app_ssh_endpoint": "some-endpoint"}` + "\n"}, nil)

				mockSession.EXPECT().Connect("some-endpoint", "cf:some-guid/0", "some-password").Return(errors.New("some error"))

				mockUI.EXPECT().Failed("Failed to connect to app over SSH:", errors.New("some error"))

				plugin.Run(mockCLI, []string{"watch", "some-app"})
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
