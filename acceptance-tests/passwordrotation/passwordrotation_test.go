package passwordrotation_test

import (
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Password Rotation", func() {
	It("should reencrypt the DB when keys are rotated", func() {
		By("creating a service broker with an encryption secret")
		firstEncryptionSecret := random.Password()
		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-rotation"),
			brokers.WithEncryptionSecrets(brokers.EncryptionSecret{
				Password: firstEncryptionSecret,
				Label:    "default",
				Primary:  true,
			}),
		)
		defer serviceBroker.Delete()

		By("creating a service")
		serviceInstance := services.CreateInstance(
			"csb-azure-storage-account",
			"standard",
			services.WithBroker(serviceBroker),
		)
		defer serviceInstance.Delete()

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
