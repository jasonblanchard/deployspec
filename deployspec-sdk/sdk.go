package deployspec

import (
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	deployspeclambda "github.com/jasonblanchard/deployspec/deployspec-sdk/lambda"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// AppSpecResourceBox these resources are polymorphic, i.e. they can be different depending on `Type`. This holds all possible values which can be checked downstream.
// Implements custom marshalling to render the appropriate type.
type AppSpecResourceBox struct {
	Type   string
	Lambda *deployspeclambda.AppSpecResource
}

func (r *AppSpecResourceBox) MarshalYAML() (interface{}, error) {
	if r.Type == "AWS::Lambda::Function" {
		return r.Lambda, nil
	}
	return nil, nil
}

func (r *AppSpecResourceBox) UnmarshalYAML(node *yaml.Node) error {
	type AppSpecResource struct {
		Type string
	}

	raw := map[string]interface{}{}
	err := node.Decode(raw)

	if err != nil {
		return err
	}

	var resource AppSpecResource
	err = mapstructure.Decode(raw, &resource)

	if err != nil {
		return err
	}

	r.Type = resource.Type

	if resource.Type == "AWS::Lambda::Function" {
		var lambdaResource deployspeclambda.AppSpecResource
		err = mapstructure.Decode(raw, &lambdaResource)
		if err != nil {
			return err
		}

		r.Lambda = &lambdaResource
	}

	return nil
}

// DeploySpecResourceBox these resources are polymorphic, i.e. they can be different depending on `Type`. This holds all possible values which can be checked downstream.
// Implements custom marshalling to render the appropriate type.
type DeploySpecResourceBox struct {
	Type   string
	Lambda *deployspeclambda.DeploySpecResource
}

func (r *DeploySpecResourceBox) UnmarshalYAML(node *yaml.Node) error {
	type DeploySpecResource struct {
		Type string
	}

	raw := map[string]interface{}{}
	err := node.Decode(raw)

	if err != nil {
		return err
	}

	var resource DeploySpecResource
	err = mapstructure.Decode(raw, &resource)

	if err != nil {
		return err
	}

	r.Type = resource.Type

	if resource.Type == "AWS::Lambda::Function" {
		var lambdaResource deployspeclambda.DeploySpecResource
		err = mapstructure.Decode(raw, &lambdaResource)
		if err != nil {
			return err
		}

		r.Lambda = &lambdaResource
	}

	return nil
}

type AppSpec struct {
	Version   string                           `yaml:"version"`
	Resources []map[string]*AppSpecResourceBox `yaml:"Resources"`
}

type DeploySpec struct {
	Version   string                              `yaml:"version"`
	Resources []map[string]*DeploySpecResourceBox `yaml:"Resources"`
	AppSpec   *AppSpec                            `yaml:"AppSpec"`
}

type Reconciler struct {
	Client *lambda.Client
}

type ReconcileOptions struct {
	DryRun bool
}

func (r *Reconciler) Reconcile(deploySpec *DeploySpec, ops *ReconcileOptions) (*AppSpec, error) {
	finalAppSpec := &AppSpec{
		Version:   deploySpec.AppSpec.Version,
		Resources: make([]map[string]*AppSpecResourceBox, 0),
	}

	deploySpecResourcesByKey := map[string]*DeploySpecResourceBox{}

	for _, deplySpecResource := range deploySpec.Resources {
		for key, resource := range deplySpecResource {
			deploySpecResourcesByKey[key] = resource
		}
	}

	for _, v := range deploySpec.AppSpec.Resources {
		for key, resource := range v {
			resourceType := resource.Type

			switch resourceType {
			case "AWS::Lambda::Function":
				reconciler := &deployspeclambda.Reconciler{
					Client: r.Client,
				}

				deploySpecResource := deploySpecResourcesByKey[key]

				finalAppspecResource, err := reconciler.ReconcileResource(resource.Lambda, deploySpecResource.Lambda, &deployspeclambda.ReconcileResourceOpts{
					Dryrun: ops.DryRun,
				})

				if err != nil {
					return nil, err
				}

				finalAppspecResourceByKey := map[string]*AppSpecResourceBox{}
				finalAppspecResourceByKey[key] = &AppSpecResourceBox{
					Type:   finalAppspecResource.Type,
					Lambda: finalAppspecResource,
				}
				finalAppSpec.Resources = append(finalAppSpec.Resources, finalAppspecResourceByKey)
			}
		}
	}

	return finalAppSpec, nil
}
