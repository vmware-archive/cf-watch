package filetree_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cf-watch/filetree"
)

var _ = Describe("FileTree", func() {
	var (
		tempDir string
		tree    *Tree
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "cf-watch")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Mkdir(filepath.Join(tempDir, "some-parent-dir"), 0755)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(tempDir, "some-parent-dir", "some-child-dir"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(tempDir, "some-parent-dir", "some-child-dir", "some-file"), []byte("some-content"), 0644)).To(Succeed())

		Expect(os.Mkdir(filepath.Join(tempDir, "some-other-parent-dir"), 0755)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(tempDir, "some-other-parent-dir", "some-child-dir"), 0755)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(tempDir, "some-other-parent-dir", "some-child-dir", "some-inaccessible-dir"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(tempDir, "some-other-parent-dir", "some-child-dir", "some-inaccessible-dir", "some-file"), []byte("some-content"), 0644)).To(Succeed())
		Expect(os.Chmod(filepath.Join(tempDir, "some-other-parent-dir", "some-child-dir", "some-inaccessible-dir"), 0000)).To(Succeed())

		tree = &Tree{}
	})

	AfterEach(func() {
		Expect(os.Chmod(filepath.Join(tempDir, "some-other-parent-dir", "some-child-dir", "some-inaccessible-dir"), 0755)).To(Succeed())
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	Describe("Tree", func() {
		Describe("#New", func() {

			It("should create a tree of nested directory structures", func() {
				file, err := tree.New(filepath.Join(tempDir, "some-parent-dir"))
				Expect(err).NotTo(HaveOccurred())
				Expect(file.Open()).To(Succeed())
				Expect(file.Basename()).To(Equal("some-parent-dir"))
				Expect(file.Children()).To(HaveLen(1))
				Expect(file.Close()).To(Succeed())

				childDir := file.Children()[0]
				Expect(childDir.Open()).To(Succeed())
				Expect(childDir.Basename()).To(Equal("some-child-dir"))
				Expect(childDir.Children()).To(HaveLen(1))
				Expect(childDir.Close()).To(Succeed())

				childFile := childDir.Children()[0]
				Expect(childFile.Open()).To(Succeed())
				Expect(childFile.Basename()).To(Equal("some-file"))
				Expect(ioutil.ReadAll(childFile)).To(Equal([]byte("some-content")))
				Expect(childFile.Children()).To(BeEmpty())
				Expect(childFile.Close()).To(Succeed())
			})

			Context("when opening a file returns an error", func() {
				It("should return an error", func() {
					_, err := tree.New("some-bad-dir")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when listing a directory returns an error", func() {
				It("should return an error", func() {
					_, err := tree.New(filepath.Join(tempDir, "some-other-parent-dir", "some-child-dir"))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when opening children of the current directory returns an error", func() {
				It("should return an error", func() {
					_, err := tree.New(filepath.Join(tempDir, "some-other-parent-dir"))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})

	Describe("File", func() {
		Describe("#Open", func() {
			FIt("should open the file", func() {
				file, err := tree.New(filepath.Join(tempDir, "some-parent-dir", "some-child-dir", "some-file"))
				Expect(err).NotTo(HaveOccurred())

				_, err := ioutil.ReadAll(file)
				Expect(err).To(MatchError("file closed"))

				Expect(file.Open()).To(Succeed())
				defer file.Close()
				Expect(ioutil.ReadAll(file)).To(Equal("some-content"))
			})
		})

		Describe("#Basename", func() {
			It("should return the base name of the file", func() {
				file, err := tree.New(filepath.Join(tempDir, "some-parent-dir", "some-child-dir", "some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(file.Basename()).To(Equal("some-file"))
			})
		})

		Describe("#Children", func() {
			It("should return the children of the file", func() {
				file, err := tree.New(filepath.Join(tempDir, "some-parent-dir"))
				Expect(err).NotTo(HaveOccurred())

				Expect(file.Children()).To(HaveLen(1))
				childDir := file.Children()[0]
				Expect(childDir.Basename()).To(Equal("some-child-dir"))
			})
		})

		Describe("#Mode", func() {
			It("should return the mode of the file", func() {
				file, err := tree.New(filepath.Join(tempDir, "some-parent-dir", "some-child-dir", "some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(file.Mode()).To(Equal(os.FileMode(0644)))
			})
		})

		Describe("#Size", func() {
			It("should return the size of the file", func() {
				file, err := tree.New(filepath.Join(tempDir, "some-parent-dir", "some-child-dir", "some-file"))
				Expect(err).NotTo(HaveOccurred())

				Expect(file.Size()).To(Equal(int64(12)))
			})
		})
	})
})
