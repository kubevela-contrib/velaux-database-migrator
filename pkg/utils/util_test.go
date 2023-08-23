package utils

import (
	"vela-migrator/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kubevela/velaux/pkg/server/infrastructure/datastore"
)

var _ = Describe("Test load_config utils", func() {
	It("Test LoadConfig function", func() {
		config, err := LoadConfig("./test_data/test_config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(config).ShouldNot(BeNil())
		cfg := types.MigratorConfig{
			Source: datastore.Config{URL: "source_url", Type: "source_type", Database: "source_db"},
			Target: datastore.Config{URL: "target_url", Type: "target_type", Database: "target_db"},
		}
		Expect(config).Should(Equal(cfg))
	})
})
