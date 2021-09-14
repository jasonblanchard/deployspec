package lambda

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

type Reconciler struct {
	Client *lambda.Client
}

type ReconcileResourceOpts struct {
	Dryrun bool
}

func (r *Reconciler) ReconcileResource(appSpecResource *AppSpecResource, deploySpecResource *DeploySpecResource, opts *ReconcileResourceOpts) (*AppSpecResource, error) {
	aliasName := "release"
	functioName := &appSpecResource.Properties.Name
	ctx := context.Background()

	// Simulate output without calling the AWS API
	if opts.Dryrun {
		output := &AppSpecResource{
			Type: appSpecResource.Type,
			Properties: &AppSpecProperties{
				Name:           appSpecResource.Properties.Name,
				Alias:          appSpecResource.Properties.Alias,
				CurrentVersion: "UNKOWN",
				TargetVersion:  "UNKOWN",
			},
		}

		return output, nil
	}

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
