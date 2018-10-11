package capi_experimental

import (
	"encoding/json"

	. "github.com/cloudfoundry/capi-bara-tests/bara_suite_helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry/capi-bara-tests/helpers/assets"
	"github.com/cloudfoundry/capi-bara-tests/helpers/random_name"
	"github.com/cloudfoundry/capi-bara-tests/helpers/services"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = CapiExperimentalDescribe("service instances", func() {
	var (
		broker               services.ServiceBroker
		serviceInstance1Name string
		serviceInstance2Name string
	)

	type ServiceInstance struct {
		Name string
	}
	type Response struct {
		Resources []ServiceInstance
	}

	BeforeEach(func() {
		broker = services.NewServiceBroker(
			random_name.BARARandomName("BRKR"),
			assets.NewAssets().ServiceBroker,
			TestSetup,
		)
		broker.Push(Config)
		broker.Configure()
		broker.Create()
		broker.PublicizePlans()

		serviceInstance1Name = random_name.BARARandomName("SVIN")
		serviceInstance2Name = random_name.BARARandomName("SVIN")

		By("Creating a service instance")
		createService := cf.Cf("create-service", broker.Service.Name, broker.SyncPlans[0].Name, serviceInstance1Name).Wait()
		Expect(createService).To(Exit(0))

		By("Creating another service instance")
		createService2 := cf.Cf("create-service", broker.Service.Name, broker.SyncPlans[0].Name, serviceInstance2Name).Wait()
		Expect(createService2).To(Exit(0))
	})

	It("Lists the service instances", func() {
		expectedResources := []ServiceInstance{
			{Name: serviceInstance1Name},
			{Name: serviceInstance2Name},
		}

		listService := cf.Cf("curl", "/v3/service_instances").Wait()
		Expect(listService).To(Exit(0))

		var res Response
		err := json.Unmarshal(listService.Out.Contents(), &res)
		Expect(err).To(BeNil())

		Expect(res.Resources).To(ConsistOf(expectedResources))
	})
})
