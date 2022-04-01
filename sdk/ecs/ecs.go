package ecs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

type AppSpecLoadBalancerInfo struct {
	ContainerName string `yaml:"ContainerName"`
	ContainerPort int    `yaml:"ContainerPort"`
}

// TODO: Build this all the way out
type AppSpecProperties struct {
	TaskDefinition   string                   `yaml:"TaskDefinition"`
	LoadBalancerInfo *AppSpecLoadBalancerInfo `yaml:"LoadBalancerInfo"`
}

type AppSpecTargetService struct {
	Type       string             `yaml:"Type"`
	Properties *AppSpecProperties `yaml:"Properties"`
}

type AppSpecResource struct {
	TargetService *AppSpecTargetService `yaml:"TargetService"`
}

type AppSpec struct {
	Resources []*AppSpecResource `yaml:"Resources"`
}

type Reconciler struct {
	Client *ecs.Client
}

type ReconcileInput struct {
	BaseAppSpec                      *AppSpec `yaml:"BaseAppSpec"`
	*ecs.RegisterTaskDefinitionInput `yaml:"RegisterTaskDefinitionInput"`
	DryRun                           bool
}

func (r *Reconciler) Reconcile(i *ReconcileInput) (*AppSpec, error) {
	ctx := context.Background()

	placeholder := "<DRY RUN PLACEHOLDER>"
	taskDefArn := &placeholder

	if !i.DryRun {
		output, err := r.Client.RegisterTaskDefinition(ctx, i.RegisterTaskDefinitionInput)

		if err != nil {
			return nil, fmt.Errorf("error registering task definition: %w", err)
		}

		taskDefArn = output.TaskDefinition.TaskDefinitionArn
	}

	appSpec := i.BaseAppSpec

	appSpec.Resources[0].TargetService.Properties.TaskDefinition = *taskDefArn

	return appSpec, nil
}
