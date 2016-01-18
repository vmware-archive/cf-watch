package file_tree_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/cf-watch/file_tree"
)

var _ = Describe("File Tree", func() {
	Describe("New", func() {
		Context("empty directory", func() {
			It("should create one node for an empty directory", func() {
				dir, err := New("fixtures/empty-dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(dir.BaseName()).To(Equal("empty-dir"))
				Expect(dir.Children()).To(BeEmpty())
			})
		})

		Context("nested directories", func() {
			It("should create nested directory structures", func() {
				dir, err := New("fixtures/nested-dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(dir.BaseName()).To(Equal("nested-dir"))
				Expect(dir.Children()).To(HaveLen(1))

				childDir := dir.Children()[0]
				Expect(childDir.BaseName()).To(Equal("child-dir"))
				Expect(childDir.Children()).To(HaveLen(1))

				childFile := childDir.Children()[0]
				Expect(childFile.BaseName()).To(Equal("some-file"))
				Expect(ioutil.ReadAll(childFile)).To(Equal([]byte("some-content")))
				Expect(childFile.Children()).To(BeEmpty())
			})
		})

		Context("when the path does not exist", func() {
			It("should return an error", func() {
				_, err := New("some-bad-dir")
				Expect(err).To(MatchError(ContainSubstring("failed to open file")))
			})
		})
	})

	Describe("#BaseName", func() {
		It("should return the base name of the file", func() {
			f, err := os.Open("fixtures/nested-dir/child-dir/some-file")
			Expect(err).NotTo(HaveOccurred())
			file := File{File: f}

			Expect(file.BaseName()).To(Equal("some-file"))
		})
	})
})
