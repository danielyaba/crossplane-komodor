/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package realtimemonitor

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	meta "github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/provider-komodor/apis/komodor/v1alpha1"
	komodorclient "github.com/crossplane/provider-komodor/internal/clients/komodor"
)

// Helper: Validate clusters
func (c *external) validateClusters(ctx context.Context, specSensors []map[string]interface{}, cr *v1alpha1.RealtimeMonitor, logger logr.Logger) error {
	var clusterNames []string
	for _, sensor := range specSensors {
		if cluster, ok := sensor["cluster"].(string); ok && cluster != "" {
			clusterNames = append(clusterNames, cluster)
		}
	}

	for _, clusterName := range clusterNames {
		logger.Info("Validating cluster", "clusterName", clusterName)
		clusterExists, err := c.client.ValidateCluster(ctx, clusterName)
		if err != nil {
			logger.Error(err, "Failed to validate cluster", "clusterName", clusterName)
			cr.SetConditions(xpv1.ReconcileError(errors.Wrapf(err, "cannot validate cluster %s", clusterName)))
			return errors.Wrapf(err, "cannot validate cluster %s", clusterName)
		}

		if !clusterExists {
			errorMsg := fmt.Sprintf("cluster '%s' does not exist in Komodor. Monitors for non-existent clusters will not be visible in the Komodor UI", clusterName)
			logger.Error(errors.New(errorMsg), "Cluster validation failed", "clusterName", clusterName)
			cr.SetConditions(xpv1.ReconcileError(errors.New(errorMsg)))
			return errors.New(errorMsg)
		}

		logger.Info("Cluster validation successful", "clusterName", clusterName)
	}
	return nil
}

// Helper: Create monitor in Komodor
func (c *external) createMonitorInKomodor(ctx context.Context, specData *specData, cr *v1alpha1.RealtimeMonitor, logger logr.Logger) (*komodorclient.Monitor, error) {
	monitor := &komodorclient.Monitor{
		Name:         cr.Spec.ForProvider.Name,
		Sensors:      specData.sensors,
		Sinks:        specData.sinks,
		Active:       cr.Spec.ForProvider.Active,
		Type:         cr.Spec.ForProvider.Type,
		Variables:    specData.variables,
		SinksOptions: cr.Spec.ForProvider.SinksOptions,
	}

	logger.Info("Sending create request to Komodor",
		"monitorName", monitor.Name,
		"monitorType", monitor.Type)

	created, err := c.client.CreateMonitor(ctx, monitor)
	if err != nil {
		logger.Error(err, "Failed to create monitor in Komodor", "monitorName", monitor.Name)
		cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot create monitor in Komodor")))
		return nil, errors.Wrap(err, "cannot create monitor in Komodor")
	}

	logger.Info("Successfully created monitor in Komodor",
		"monitorID", created.ID,
		"monitorName", created.Name)

	return created, nil
}

// Helper: Update resource from created monitor
func (c *external) updateResourceFromCreatedMonitor(cr *v1alpha1.RealtimeMonitor, created *komodorclient.Monitor, logger logr.Logger) {
	// Set external-name to the Komodor monitor ID
	meta.SetExternalName(cr, created.ID)

	if err := updateStatusFromMonitor(cr, created); err != nil {
		logger.Error(err, "Failed to update status from created monitor")
		return
	}

	cr.SetConditions(xpv1.Creating(), xpv1.ReconcileSuccess())
	logger.Info("Create completed successfully", "monitorID", created.ID)
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	logger := log.FromContext(ctx)

	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRealtimeMonitor)
	}

	logger.Info("Creating RealtimeMonitor",
		"name", cr.Name,
		"namespace", cr.Namespace,
		"monitorName", cr.Spec.ForProvider.Name)

	// Unmarshal spec data
	specData, err := unmarshalSpecData(cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	// Validate clusters
	if err := c.validateClusters(ctx, specData.sensors, cr, logger); err != nil {
		return managed.ExternalCreation{}, err
	}

	// Create monitor in Komodor
	logger.Info("Proceeding with monitor creation", "monitorName", cr.Spec.ForProvider.Name)
	created, err := c.createMonitorInKomodor(ctx, specData, cr, logger)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	// Update resource with created monitor data
	c.updateResourceFromCreatedMonitor(cr, created, logger)

	return managed.ExternalCreation{}, nil
}
