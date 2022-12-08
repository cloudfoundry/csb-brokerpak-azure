package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"
)

func search(id string) (ServicePlansData, error) {
	url := fmt.Sprintf("/v3/service_plans?broker_catalog_ids=%s", id)
	session := cf.Start("curl", url)
	Eventually(session, time.Minute).Should(gexec.Exit(0))
	err := string(session.Err.Contents())
	if err != "" {
		return ServicePlansData{}, errors.New(err)
	}

	var data ServicePlansData
	if err := json.NewDecoder(session.Out).Decode(&data); err != nil {
		return ServicePlansData{}, err
	}

	return data, nil
}

type ResourceData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ServicePlansData struct {
	Resources []ResourceData `json:"resources"`
}

type servicePlanNameByIDSearcher struct {
	searcherFn func(id string) (ServicePlansData, error)
}

func NewServicePlanNameByIDSearcher(searcherFn func(id string) (ServicePlansData, error)) *servicePlanNameByIDSearcher {
	if searcherFn == nil {
		searcherFn = search
	}
	return &servicePlanNameByIDSearcher{searcherFn: searcherFn}
}

func (s servicePlanNameByIDSearcher) Search(planID string) (string, error) {
	servicePlanData, err := s.searcherFn(planID)
	if err != nil {
		return "", err
	}

	for _, resource := range servicePlanData.Resources {
		return resource.Name, nil
	}
	return "", fmt.Errorf("plan by id %s not found", planID)
}
