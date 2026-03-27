package catalog

import (
	"os"

	"github.com/keycloak-broker/pkg/logger"
	"gopkg.in/yaml.v3"
)

var catalog Catalog

type Catalog struct {
	Services []Service `yaml:"services"`
}

type Service struct {
	ID                   string          `yaml:"id" json:"id"`
	Name                 string          `yaml:"name" json:"name"`
	Description          string          `yaml:"description" json:"description"`
	Bindable             bool            `yaml:"bindable" json:"bindable"`
	InstancesRetrievable bool            `yaml:"instances_retrievable" json:"instances_retrievable"`
	BindingsRetrievable  bool            `yaml:"bindings_retrievable" json:"bindings_retrievable"`
	PlanUpdateable       bool            `yaml:"plan_updateable" json:"plan_updateable"`
	Tags                 []string        `yaml:"tags" json:"tags,omitempty"`
	Metadata             ServiceMetadata `yaml:"metadata" json:"metadata"`
	Plans                []Plan          `yaml:"plans" json:"plans"`
}

type ServiceMetadata struct {
	DisplayName         string `yaml:"displayName" json:"displayName"`
	ImageUrl            string `yaml:"imageUrl" json:"imageUrl,omitempty"`
	LongDescription     string `yaml:"longDescription" json:"longDescription,omitempty"`
	ProviderDisplayName string `yaml:"providerDisplayName" json:"providerDisplayName,omitempty"`
	DocumentationUrl    string `yaml:"documentationUrl" json:"documentationUrl,omitempty"`
	SupportUrl          string `yaml:"supportUrl" json:"supportUrl,omitempty"`
}

type Plan struct {
	ID          string       `yaml:"id" json:"id"`
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Free        bool         `yaml:"free" json:"free,omitempty"`
	Metadata    PlanMetadata `yaml:"metadata" json:"metadata"`
}

type PlanMetadata struct {
	PublicClient        bool `yaml:"publicClient" json:"publicClient,omitempty"`
}

func init() {
	data, err := os.ReadFile("catalog.yaml")
	if err != nil {
		logger.Fatal("failed to read catalog.yaml: %v", err)
	}
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		logger.Fatal("failed to parse catalog.yaml: %v", err)
	}
	logger.Info("loaded catalog with %d service(s)", len(catalog.Services))
}

func GetCatalog() map[string]any {
	return map[string]any{"services": catalog.Services}
}

func IsPublicClient(serviceID, planID string) bool {
	for _, svc := range catalog.Services {
		if svc.ID == serviceID {
			for _, p := range svc.Plans {
				if p.ID == planID {
					return p.Metadata.PublicClient
				}
			}
		}
	}
	return false
}

func GetPlan(serviceID, planID string) Plan {
	for _, svc := range catalog.Services {
		if svc.ID == serviceID {
			for _, p := range svc.Plans {
				if p.ID == planID {
					return p
				}
			}
		}
	}
	return Plan{}
}
