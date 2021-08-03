package cosmosdb_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCosmosDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CosmosDB Suite")
}
