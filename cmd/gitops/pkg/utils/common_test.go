package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test common utils", func() {
	Describe("FindCoreConfig", func() {
		var dir string

		BeforeEach(func() {
			var err error
			dir, err = os.MkdirTemp("", "find-core-config")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		It("fails to find core configuration in empty dir", func() {
			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Missing, Path: ""}))
		})

		It("fails to find core configuration in non-empty dir with no config file", func() {
			configFile, err := ioutil.ReadFile("testdata/ingress.yaml")
			Expect(err).NotTo(HaveOccurred())

			path := filepath.Join(dir, "ingress.yaml")
			Expect(ioutil.WriteFile(path, configFile, 0666)).To(Succeed())

			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Missing, Path: ""}))
		})

		It("finds embedded core configuration in dir containing file with wrong number of entries", func() {
			configFile, err := ioutil.ReadFile("testdata/config.yaml")
			Expect(err).NotTo(HaveOccurred())

			// Add an extraneous entry
			configFile = append(configFile, []byte("---\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: my-ns\n")...)
			path := filepath.Join(dir, "config.yaml")
			Expect(ioutil.WriteFile(path, configFile, 0666)).To(Succeed())

			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Embedded, Path: path}))
		})

		It("finds partial core configuration in dir containing file with partial config", func() {
			partialConfigFile, err := ioutil.ReadFile("testdata/partial-config.yaml")
			Expect(err).NotTo(HaveOccurred())

			path := filepath.Join(dir, "config.yaml")
			Expect(ioutil.WriteFile(path, partialConfigFile, 0666)).To(Succeed())

			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Embedded, Path: path}))
		})

		It("finds core configuration nested in dir containing one (regardless of name)", func() {
			testConfigFile, err := ioutil.ReadFile("testdata/config.yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(os.MkdirAll(filepath.Join(dir, "nested"), 0700)).Should(Succeed())
			renamedConfigFile := filepath.Join(dir, "nested", "sprug.yaml")
			Expect(ioutil.WriteFile(renamedConfigFile, testConfigFile, 0666)).To(Succeed())

			Expect(FindCoreConfig(dir)).To(Equal(WalkResult{Status: Valid, Path: renamedConfigFile}))
		})
	})
})
