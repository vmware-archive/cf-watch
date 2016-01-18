package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var context struct {
	cfDomain   string
	cfUsername string
	cfPassword string
	cfHome     string
}

var _ = BeforeSuite(func() {
	context.cfDomain = loadEnv("CF_DOMAIN")
	context.cfUsername = loadEnv("CF_USERNAME")
	context.cfPassword = loadEnv("CF_PASSWORD")

	var err error
	context.cfHome, err = ioutil.TempDir("", "cf-watch")
	Expect(err).NotTo(HaveOccurred())
	os.Setenv("CF_HOME", context.cfHome)
	os.Setenv("CF_PLUGIN_HOME", filepath.Join(context.cfHome, "plugins"))

	pluginPath, err := gexec.Build("github.com/pivotal-cf/cf-watch")
	Expect(err).NotTo(HaveOccurred())

	Eventually(cf("install-plugin", "-f", pluginPath)).Should(gexec.Exit(0))
})

var _ = AfterSuite(func() {
	Expect(os.RemoveAll(context.cfHome)).To(Succeed())
})

var _ = Describe("CF Watch", func() {
	var orgName string

	BeforeEach(func() {
		Eventually(cf("api", "api."+context.cfDomain, "--skip-ssl-validation")).Should(gexec.Exit(0))
		Eventually(cf("auth", context.cfUsername, context.cfPassword)).Should(gexec.Exit(0))
		orgUUID, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())
		orgName = fmt.Sprint("org-", orgUUID)
		Eventually(cf("create-org", orgName), "2s").Should(gexec.Exit(0))
		Eventually(cf("create-space", "test-space", "-o", orgName)).Should(gexec.Exit(0))
		Eventually(cf("target", "-o", orgName, "-s", "test-space")).Should(gexec.Exit(0))
		Eventually(cf("push", "test-app", "-p", "fixtures/test-app", "-b", "go_buildpack", "--no-route"), "2m").Should(gexec.Exit(0))
	})

	AfterEach(func() {
		Eventually(cf("delete-org", "-f", orgName), "10s").Should(gexec.Exit(0))
	})

	It("should write a `/tmp/watch` file to the app container", func() {
		Eventually(cf("watch", "test-app", "fixtures/some-dir/some-nested-dir/some-file"), "2s").Should(gexec.Exit(0))
		session := cf("ssh", "test-app", "-k", "-c", "cat /tmp/watch", "-i", "0")
		Eventually(session).Should(gexec.Exit(0))
		Expect(session).To(gbytes.Say("some-text"))
	})
})

func loadEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		Fail("missing "+name, 1)
	}
	return value
}

func cf(args ...string) *gexec.Session {
	command := exec.Command("cf", args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return session
}
