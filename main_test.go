package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = BeforeSuite(func() {
	uninstallCommand := exec.Command("cf", "uninstall-plugin", "Watch")
	session, err := gexec.Start(uninstallCommand, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	pluginPath, err := gexec.Build("github.com/sclevine/cf-watch")
	Expect(err).NotTo(HaveOccurred())
	installCommand := exec.Command("cf", "install-plugin", "-f", pluginPath)
	session, err = gexec.Start(installCommand, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
})

var _ = AfterSuite(func() {
	uninstallCommand := exec.Command("cf", "uninstall-plugin", "Watch")
	session, err := gexec.Start(uninstallCommand, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
})

var _ = Describe("CF Watch", func() {
	It("should print watch", func() {
		watchCommand := exec.Command("cf", "watch")
		session, err := gexec.Start(watchCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(session).To(gbytes.Say("watch"))
	})
})
