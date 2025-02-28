package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/cloudfoundry/cf-test-helpers/v2/helpers"
	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"
	bara_config "github.com/cloudfoundry/capi-bara-tests/helpers/config"
	"github.com/cloudfoundry/capi-bara-tests/helpers/v3_helpers"

	. "github.com/cloudfoundry/capi-bara-tests/bara_suite_helpers"
	"github.com/cloudfoundry/capi-bara-tests/helpers/assets"
	"github.com/cloudfoundry/capi-bara-tests/helpers/random_name"
)

type Plan struct {
	Name    string      `json:"name"`
	ID      string      `json:"id"`
	Schemas PlanSchemas `json:"schemas"`
}

type PlanSchemas struct {
	ServiceInstance struct {
		Create struct {
			Parameters map[string]interface{} `json:"parameters"`
		} `json:"create"`
		Update struct {
			Parameters map[string]interface{} `json:"parameters"`
		} `json:"update"`
	} `json:"service_instance"`
	ServiceBinding struct {
		Create struct {
			Parameters map[string]interface{} `json:"parameters"`
		} `json:"create"`
	} `json:"service_binding"`
}

type ServiceBroker struct {
	Name       string
	SpaceGUID  string
	DomainGUID string
	Path       string
	TestSetup  *workflowhelpers.ReproducibleTestSuiteSetup
	Service    struct {
		Name            string `json:"name"`
		ID              string `json:"id"`
		DashboardClient struct {
			ID          string `json:"id"`
			Secret      string `json:"secret"`
			RedirectUri string `json:"redirect_uri"`
		}
	}
	SyncPlans  []Plan
	AsyncPlans []Plan
}

type ServicesResponse struct {
	Resources []ServiceResponse
}

type ServiceResponse struct {
	Entity struct {
		Label        string
		ServicePlans []ServicePlanResponse `json:"service_plans"`
	}
}

type ServicePlansResponse struct {
	Resources []ServicePlanResponse
}

type ServicePlanResponse struct {
	Entity struct {
		Name    string
		Public  bool
		Schemas PlanSchemas
	}
	Metadata struct {
		Url  string
		Guid string
	}
}

type ServiceInstance struct {
	Metadata struct {
		Guid string `json:"guid"`
	}
}

type ServiceInstanceResponse struct {
	Resources []ServiceInstance
}

type SpaceJson struct {
	Resources []struct {
		Metadata struct {
			Guid string
		}
	}
}

func NewServiceBroker(name, spaceGUID, domainGUID, path string, TestSetup *workflowhelpers.ReproducibleTestSuiteSetup) ServiceBroker {
	b := ServiceBroker{}
	b.Path = path
	b.Name = name
	b.SpaceGUID = spaceGUID
	b.DomainGUID = domainGUID
	b.Service.Name = random_name.BARARandomName("SVC")
	b.Service.ID = random_name.BARARandomName("SVC-ID")

	b.SyncPlans = []Plan{
		{Name: random_name.BARARandomName("SVC-PLAN"), ID: random_name.BARARandomName("SVC-PLAN-ID")},
		{Name: random_name.BARARandomName("SVC-PLAN"), ID: random_name.BARARandomName("SVC-PLAN-ID")},
	}
	b.AsyncPlans = []Plan{
		{Name: random_name.BARARandomName("SVC-PLAN"), ID: random_name.BARARandomName("SVC-PLAN-ID")},
		{Name: random_name.BARARandomName("SVC-PLAN"), ID: random_name.BARARandomName("SVC-PLAN-ID")},
		{Name: random_name.BARARandomName("SVC-PLAN"), ID: random_name.BARARandomName("SVC-PLAN-ID")},
	}
	b.Service.DashboardClient.ID = random_name.BARARandomName("DASHBOARD-ID")
	b.Service.DashboardClient.Secret = random_name.BARARandomName("DASHBOARD-SECRET")
	b.Service.DashboardClient.RedirectUri = random_name.BARARandomName("DASHBOARD-URI")
	b.TestSetup = TestSetup
	return b
}

func (b ServiceBroker) Push(config bara_config.BaraConfig) {
	appGUID := v3_helpers.CreateApp(b.Name, b.SpaceGUID, `{}`)
	v3_helpers.CreateAndMapRoute(appGUID, b.SpaceGUID, b.DomainGUID, b.Name)
	Expect(cf.Cf("set-health-check", b.Name, "http", "--endpoint", "/v2/catalog").Wait(Config.BrokerStartTimeoutDuration())).To(Exit(0))
	Expect(cf.Cf(
		"push", b.Name,
		"-b", config.GetRubyBuildpackName(),
		"-m", DEFAULT_MEMORY_LIMIT,
		"-p", b.Path,
	).Wait(Config.BrokerStartTimeoutDuration())).To(Exit(0))
}

func (b ServiceBroker) Configure() {
	Expect(helpers.Curl(Config, helpers.AppUri(b.Name, "/config", Config), "-d", b.ToJSON()).Wait()).To(Exit(0))
}

func (b ServiceBroker) Restart() {
	Expect(cf.Cf("restart", b.Name).Wait(Config.BrokerStartTimeoutDuration())).To(Exit(0))
}

func (b ServiceBroker) Create() {
	workflowhelpers.AsUser(b.TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
		Expect(cf.Cf("create-service-broker", b.Name, "username", "password", helpers.AppUri(b.Name, "", Config)).Wait()).To(Exit(0))
		Expect(cf.Cf("service-brokers").Wait()).To(Say(b.Name))
	})
}

func (b ServiceBroker) CreateSpaceScoped() {
	workflowhelpers.AsUser(b.TestSetup.RegularUserContext(), Config.DefaultTimeoutDuration(), func() {
		Expect(cf.Cf("create-service-broker", b.Name, "username", "password", helpers.AppUri(b.Name, "", Config), "--space-scoped").Wait()).To(Exit(0))
		Expect(cf.Cf("service-brokers").Wait()).To(Say(b.Name))
	})
}

func (b ServiceBroker) Update() {
	workflowhelpers.AsUser(b.TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
		Expect(cf.Cf("update-service-broker", b.Name, "username", "password", helpers.AppUri(b.Name, "", Config)).Wait()).To(Exit(0))
	})
}

func (b ServiceBroker) Delete() {
	workflowhelpers.AsUser(b.TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
		Expect(cf.Cf("delete-service-broker", b.Name, "-f").Wait()).To(Exit(0))

		brokers := cf.Cf("service-brokers").Wait()
		Expect(brokers).To(Exit(0))
		Expect(brokers.Out.Contents()).ToNot(ContainSubstring(b.Name))
	})
}

func (b ServiceBroker) Destroy() {
	workflowhelpers.AsUser(b.TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
		Expect(cf.Cf("purge-service-offering", b.Service.Name, "-f").Wait()).To(Exit(0))
	})
	b.Delete()
	Expect(cf.Cf("delete", b.Name, "-f", "-r").Wait()).To(Exit(0))
}

func (b ServiceBroker) ToJSON() string {
	bytes, err := ioutil.ReadFile(assets.NewAssets().ServiceBroker + "/cats.json")
	Expect(err).To(BeNil())

	planSchema, err := json.Marshal(b.SyncPlans[0].Schemas)
	Expect(err).To(BeNil())

	replacer := strings.NewReplacer(
		"<fake-service>", b.Service.Name,
		"<fake-service-guid>", b.Service.ID,
		"<sso-test>", b.Service.DashboardClient.ID,
		"<sso-secret>", b.Service.DashboardClient.Secret,
		"<sso-redirect-uri>", b.Service.DashboardClient.RedirectUri,
		"<fake-plan>", b.SyncPlans[0].Name,
		"<fake-plan-guid>", b.SyncPlans[0].ID,
		"<fake-plan-2>", b.SyncPlans[1].Name,
		"<fake-plan-2-guid>", b.SyncPlans[1].ID,
		"<fake-async-plan>", b.AsyncPlans[0].Name,
		"<fake-async-plan-guid>", b.AsyncPlans[0].ID,
		"<fake-async-plan-2>", b.AsyncPlans[1].Name,
		"<fake-async-plan-2-guid>", b.AsyncPlans[1].ID,
		"<fake-async-plan-3>", b.AsyncPlans[2].Name,
		"<fake-async-plan-3-guid>", b.AsyncPlans[2].ID,
		"\"<fake-plan-schema>\"", string(planSchema),
	)

	return replacer.Replace(string(bytes))
}

func (b ServiceBroker) PublicizePlans() {
	url := fmt.Sprintf("/v2/services?inline-relations-depth=1&q=label:%s", b.Service.Name)
	var session *Session
	workflowhelpers.AsUser(b.TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
		session = cf.Cf("curl", url).Wait()
		Expect(session).To(Exit(0))
	})
	structure := ServicesResponse{}
	json.Unmarshal(session.Out.Contents(), &structure)

	for _, service := range structure.Resources {
		if service.Entity.Label == b.Service.Name {
			for _, plan := range service.Entity.ServicePlans {
				if b.HasPlan(plan.Entity.Name) {
					b.PublicizePlan(plan.Metadata.Url)
				}
			}
		}
	}
}

func (b ServiceBroker) HasPlan(planName string) bool {
	for _, plan := range b.Plans() {
		if plan.Name == planName {
			return true
		}
	}
	return false
}

func (b ServiceBroker) PublicizePlan(url string) {
	jsonMap := make(map[string]bool)
	jsonMap["public"] = true
	planJson, _ := json.Marshal(jsonMap)
	workflowhelpers.AsUser(b.TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
		Expect(cf.Cf("curl", url, "-X", "PUT", "-d", string(planJson)).Wait()).To(Exit(0))
	})
}

func (b ServiceBroker) CreateServiceInstance(instanceName string) string {
	Expect(cf.Cf("create-service", b.Service.Name, b.SyncPlans[0].Name, instanceName).Wait()).To(Exit(0))
	url := fmt.Sprintf("/v2/service_instances?q=name:%s", instanceName)
	serviceInstance := ServiceInstanceResponse{}
	curl := cf.Cf("curl", url).Wait()
	Expect(curl).To(Exit(0))
	json.Unmarshal(curl.Out.Contents(), &serviceInstance)
	return serviceInstance.Resources[0].Metadata.Guid
}

func (b ServiceBroker) Plans() []Plan {
	plans := make([]Plan, 0)
	plans = append(plans, b.SyncPlans...)
	plans = append(plans, b.AsyncPlans...)
	return plans
}
