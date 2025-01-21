package acceptance_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
)

// Tests rotation of the encryption password using the *csb-azure-mongodb* service offering
// Does NOT use the default broker: deploys a custom-configured broker
var _ = Describe("Password Rotation", Label("passwordrotation"), func() {
	It("should reencrypt the DB when keys are rotated", func() {
		By("creating a service broker with an encryption secret")
		firstEncryptionSecret := random.Password()
		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-rotation"),
			brokers.WithLatestEnv(),
			brokers.WithEncryptionSecrets(brokers.EncryptionSecret{
				Password: firstEncryptionSecret,
				Label:    "default",
				Primary:  true,
			}),
		)
		defer serviceBroker.Delete()

		By("creating a service")
		databaseName := random.Name(random.WithPrefix("database"))
		collectionName := random.Name(random.WithPrefix("collection"))

		const serviceOffering = "csb-azure-mongodb"
		const servicePlan = "small"
		serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
		// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
		// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
		// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
		// (which can only have 5 instances at any time) to prevent subsequent test failures.
		defer services.Delete(serviceName)
		serviceInstance := services.CreateInstance(
			serviceOffering,
			servicePlan,
			services.WithBroker(serviceBroker),
			services.WithParameters(map[string]any{
				"db_name":         databaseName,
				"collection_name": collectionName,
				"shard_key":       "_id",
				"indexes":         "_id",
				"unique_indexes":  "",
			}),
			services.WithName(serviceName),
		)

		By("adding a new encryption secret")
		secondEncryptionSecret := random.Password()
		serviceBroker.UpdateEncryptionSecrets(
			brokers.EncryptionSecret{
				Password: firstEncryptionSecret,
				Label:    "default",
				Primary:  false,
			},
			brokers.EncryptionSecret{
				Password: secondEncryptionSecret,
				Label:    "second-password",
				Primary:  true,
			},
		)

		By("creating a service key")
		sk1 := serviceInstance.CreateServiceKey()
		defer sk1.Delete()

		By("removing the original encryption secret")
		serviceBroker.UpdateEncryptionSecrets(
			brokers.EncryptionSecret{
				Password: secondEncryptionSecret,
				Label:    "second-password",
				Primary:  true,
			},
		)

		By("creating a new service key")
		sk2 := serviceInstance.CreateServiceKey()
		defer sk2.Delete()
	})
})
