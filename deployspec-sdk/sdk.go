package deployspec

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type DeploySpecFunctionCode struct {
	S3Bucket string `yaml:"S3Bucket"`
	S3Key    string `yaml:"S3Key"`
}

type DeploySpecFunctionConfiguration struct {
	Environment map[string]string `yaml:"Environment"`
}

type DeploySpecResource struct {
	FunctionCode                    *DeploySpecFunctionCode          `yaml:"FunctionCode"`
	DeploySpecFunctionConfiguration *DeploySpecFunctionConfiguration `yaml:"FunctionConfiguration"`
}

type AppSpecProperties struct {
	Name           string `yaml:"Name"`
	Alias          string `yaml:"Alias"`
	CurrentVersion string `yaml:"CurrentVersion"`
	TargetVersion  string `yaml:"TargetVersion"`
}

type AppSpecResource struct {
	Type       string             `yaml:"Type"`
	Properties *AppSpecProperties `yaml:"Properties"`
}

type AppSpec struct {
	Version   string                        `yaml:"version"`
	Resources []map[string]*AppSpecResource `yaml:"Resources"`
}

type DeploySpec struct {
	Version   string                           `yaml:"version"`
	Resources []map[string]*DeploySpecResource `yaml:"Resources"`
	AppSpec   *AppSpec                         `yaml:"AppSpec"`
}

type Reconciler struct {
	Client *lambda.Client
}

func (r *Reconciler) Reconcile(deploySpec *DeploySpec) (*AppSpec, error) {
	finalAppSpec := &AppSpec{
		Version:   deploySpec.AppSpec.Version,
		Resources: make([]map[string]*AppSpecResource, 0),
	}

	deploySpecResourcesByKey := map[string]DeploySpecResource{}

	for _, deplySpecResource := range deploySpec.Resources {
		for key, resource := range deplySpecResource {
			deploySpecResourcesByKey[key] = *resource
		}
	}

	for _, v := range deploySpec.AppSpec.Resources {
		for key, resource := range v {
			resourceType := resource.Type

			switch resourceType {
			case "AWS::Lambda::Function":
				reconciler := &LambdaReconciler{
					Client: r.Client,
				}

				deploySpecResource := deploySpecResourcesByKey[key]

				finalAppspecResource, err := reconciler.ReconcileResource(resource, &deploySpecResource)

				// TOOD: This error is getting swallowed
				if err != nil {
					return nil, err
				}

				finalAppspecResourceByKey := map[string]*AppSpecResource{}
				finalAppspecResourceByKey[key] = finalAppspecResource
				finalAppSpec.Resources = append(finalAppSpec.Resources, finalAppspecResourceByKey)
			}
		}
	}

	return finalAppSpec, nil
}

type LambdaReconciler struct {
	Client *lambda.Client
}

func (r *LambdaReconciler) ReconcileResource(appSpecResource *AppSpecResource, deploySpecResource *DeploySpecResource) (*AppSpecResource, error) {
	aliasName := "release"
	functioName := &appSpecResource.Properties.Name
	ctx := context.Background()

	alias, err := r.Client.GetAlias(ctx, &lambda.GetAliasInput{
		FunctionName: functioName,
		Name:         &aliasName,
	})

	if err != nil {
		return nil, err
	}

	currentVersion := *alias.FunctionVersion

	updateFunctionCodeOutput, err := r.Client.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: functioName,
		S3Bucket:     &deploySpecResource.FunctionCode.S3Bucket,
		S3Key:        &deploySpecResource.FunctionCode.S3Key,
	})

	if err != nil {
		return nil, err
	}

	revisionId := *updateFunctionCodeOutput.RevisionId
	codeSha := *updateFunctionCodeOutput.CodeSha256

	// TODO: Merge this with current environment
	// environment := &lambdatypes.Environment{
	// 	Variables: deploySpecResource.DeploySpecFunctionConfiguration.Environment,
	// }

	// updateFunctionConfigurationOutput, err := r.Client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
	// 	FunctionName: functioName,
	// 	RevisionId:   &revisionId,
	// 	Environment:  environment,
	// })

	// if err != nil {
	// 	return nil, err
	// }

	// revisionId = *updateFunctionConfigurationOutput.RevisionId
	// codeSha = *updateFunctionConfigurationOutput.CodeSha256

	description := fmt.Sprintf("s3://%s/%s", deploySpecResource.FunctionCode.S3Bucket, deploySpecResource.FunctionCode.S3Key)

	publishVersionOutput, err := r.Client.PublishVersion(ctx, &lambda.PublishVersionInput{
		FunctionName: functioName,
		RevisionId:   &revisionId,
		CodeSha256:   &codeSha,
		Description:  &description,
	})

	if err != nil {
		return nil, err
	}

	targetVersion := *publishVersionOutput.Version

	// Return full AppSpec with currentVersion & targetVersion set
	output := &AppSpecResource{
		Type: appSpecResource.Type,
		Properties: &AppSpecProperties{
			Name:           appSpecResource.Properties.Name,
			Alias:          appSpecResource.Properties.Alias,
			CurrentVersion: currentVersion,
			TargetVersion:  targetVersion,
		},
	}

	return output, nil
}
