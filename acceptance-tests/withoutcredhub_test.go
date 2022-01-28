package acceptance_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", Label("withoutcredhub"), func() {
	It("can be accessed by an app", func() {
		env := apps.EnvVar{Name: "CH_CRED_HUB_URL", Value: ""}
		broker := brokers.Create(
			brokers.WithPrefix("csb-storage"),
			brokers.WithEnv(env),
		)
		defer broker.Delete()

		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-azure-storage-account",
			"standard",
			services.WithBroker(broker),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		app := apps.Push(apps.WithApp(apps.Storage))
		defer apps.Delete(app)

		By("binding the app to the storage service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("checking that the app environment does not a credhub reference for credentials")
		Expect(binding.Credential()).NotTo(matchers.HaveCredHubRef)

		By("creating a collection")
		collectionName := random.Name(random.WithPrefix("collection"))
		app.PUT("", collectionName)

		By("uploading a blob")
		blobName := random.Hexadecimal()
		blobData := random.Hexadecimal()
		app.PUT(blobData, "%s/%s", collectionName, blobName)

		By("downloading the blob")
		got := app.GET("%s/%s", collectionName, blobName)
		Expect(got).To(Equal(blobData))
	})
})
