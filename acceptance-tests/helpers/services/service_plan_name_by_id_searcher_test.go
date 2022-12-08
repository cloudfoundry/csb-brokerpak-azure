package services_test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakazure/acceptance-tests/helpers/services"
)

var _ = Describe("Service Plan Name By ID Searcher", Label("services"), func() {
	When("searching a service plan name by id", func() {
		It("should returns the name", func() {
			planID := "6b9ca24e-1dec-4e6f-8c8a-dc6e11ab5bef"
			planName, err := services.NewServicePlanNameByIDSearcher(fakeSearcher).Search(planID)
			Expect(err).NotTo(HaveOccurred())
			Expect(planName).To(Equal("small"))
		})
	})
})

func fakeSearcher(_ string) (services.ServicePlansData, error) {
	absPath, err := filepath.Abs("./broker_catalog_ids.json")
	if err != nil {
		return services.ServicePlansData{}, err
	}
	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return services.ServicePlansData{}, err
	}

	var data services.ServicePlansData
	if err := json.Unmarshal(content, &data); err != nil {
		return services.ServicePlansData{}, err
	}

	return data, nil
}
