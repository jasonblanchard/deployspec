package ecs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"gotest.tools/assert"
)

type mockClient struct{}

func (c *mockClient) RegisterTaskDefinition(ctx context.Context, params *ecs.RegisterTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.RegisterTaskDefinitionOutput, error) {
	arn := "MockArn"
	output := &ecs.RegisterTaskDefinitionOutput{
		TaskDefinition: &ecstypes.TaskDefinition{
			TaskDefinitionArn: &arn,
		},
	}
	return output, nil
}

func TestReconcile(t *testing.T) {
	reconciler := &Reconciler{
		Client: &mockClient{},
	}

	input := &ReconcileInput{
		BaseAppSpec: &AppSpec{
			Resources: []*AppSpecResource{
				{
					TargetService: &AppSpecTargetService{
						Properties: &AppSpecProperties{},
					},
				},
			},
		},
	}

	result, err := reconciler.Reconcile(input)

	assert.NilError(t, err)

	assert.Equal(t, result.Resources[0].TargetService.Properties.TaskDefinition, "MockArn")
}

func TestReconcileDryRun(t *testing.T) {
	reconciler := &Reconciler{
		Client: &mockClient{},
	}

	input := &ReconcileInput{
		BaseAppSpec: &AppSpec{
			Resources: []*AppSpecResource{
				{
					TargetService: &AppSpecTargetService{
						Properties: &AppSpecProperties{},
					},
				},
			},
		},
		DryRun: true,
	}

	result, err := reconciler.Reconcile(input)

	assert.NilError(t, err)

	assert.Equal(t, result.Resources[0].TargetService.Properties.TaskDefinition, "<DRY RUN PLACEHOLDER>")
}
