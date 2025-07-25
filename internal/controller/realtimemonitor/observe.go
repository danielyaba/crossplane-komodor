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
	"strings"

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

// Helper: Fetch monitor from Komodor
func (c *external) fetchMonitorFromKomodor(ctx context.Context, monitorID string) (*komodorclient.Monitor, error) {
	logger := log.FromContext(ctx)
	logger.Info("Fetching monitor from Komodor", "monitorID", monitorID)

	monitor, err := c.client.GetMonitor(ctx, monitorID)
	if err != nil {
		logger.Error(err, "Failed to get monitor from Komodor", "monitorID", monitorID)
		return nil, err
	}

	logger.Info("Successfully fetched monitor from Komodor",
		"monitorID", monitorID,
		"monitorName", monitor.Name,
		"isDeleted", monitor.IsDeleted)

	return monitor, nil
}

// Helper: Set observe conditions
func (c *external) setObserveConditions(cr *v1alpha1.RealtimeMonitor, resourceUpToDate bool, monitorID string, logger logr.Logger) {
	if resourceUpToDate {
		cr.SetConditions(xpv1.Available(), xpv1.ReconcileSuccess())
		logger.Info("Monitor is up-to-date, set READY and SYNCED conditions to True", "monitorID", monitorID)
	} else {
		cr.SetConditions(xpv1.Available())
		logger.Info("Monitor exists but needs update, set READY condition to True", "monitorID", monitorID)
	}
}

// Helper: Handle error from Komodor GetMonitor
func handleGetMonitorError(ctx context.Context, cr *v1alpha1.RealtimeMonitor, extName string, err error) (managed.ExternalObservation, error) {
	logger := log.FromContext(ctx)

	// Check if this is a 404 Not Found error
	if komodorclient.IsNotFound(err) {
		logger.Info("Monitor not found in Komodor (404)", "monitorID", extName)
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Check if this is a 400 Bad Request or 403 Forbidden with an invalid external name (not a UUID)
	// This happens when Crossplane automatically sets external-name to the Kubernetes resource name
	if strings.Contains(err.Error(), "400 Bad Request") || strings.Contains(err.Error(), "403 Forbidden") {
		if !isValidUUID(extName) {
			logger.Info("400/403 error with invalid external name (not UUID), clearing external name to trigger creation",
				"externalName", extName,
				"resourceName", cr.Name,
				"error", err.Error())
			// Clear the incorrect external name to allow recreation
			meta.SetExternalName(cr, "")
			return managed.ExternalObservation{ResourceExists: false}, nil
		} else {
			// Valid UUID but still getting 400/403 - this might be a different issue
			logger.Error(err, "400/403 error with valid UUID - this might indicate a different issue",
				"externalName", extName,
				"resourceName", cr.Name)
		}
	}

	// For other errors, set reconcile error condition
	logger.Error(err, "Unexpected error getting monitor from Komodor",
		"externalName", extName,
		"resourceName", cr.Name)
	cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot get monitor from Komodor")))
	return managed.ExternalObservation{}, errors.Wrapf(err, "failed to get monitor %q from Komodor", extName)
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	logger := log.FromContext(ctx)

	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRealtimeMonitor)
	}

	logger.Info("Observing RealtimeMonitor",
		"name", cr.Name,
		"namespace", cr.Namespace,
		"externalName", meta.GetExternalName(cr))

	// Check if monitor exists
	monitorID := meta.GetExternalName(cr)
	if monitorID == "" {
		logger.Info("No external name found, resource does not exist")
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Validate external name format
	if !isValidUUID(monitorID) {
		logger.Info("External name is not a valid UUID, clearing it to trigger creation",
			"externalName", monitorID,
			"resourceName", cr.Name)
		meta.SetExternalName(cr, "")
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Fetch monitor from Komodor
	monitor, err := c.fetchMonitorFromKomodor(ctx, monitorID)
	if err != nil {
		return handleGetMonitorError(ctx, cr, monitorID, err)
	}

	// Check if monitor is deleted
	if monitor.IsDeleted {
		logger.Info("Monitor is marked as deleted, treating as non-existent", "monitorID", monitorID)
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Unmarshal spec data
	specData, err := unmarshalSpecData(cr)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// Check if monitor is up to date
	resourceUpToDate := isMonitorUpToDate(&cr.Spec.ForProvider, monitor, specData.sensors, specData.sinks, specData.variables)
	logger.Info("Monitor comparison completed",
		"monitorID", monitorID,
		"resourceUpToDate", resourceUpToDate)

	// Update status from monitor
	if err := updateStatusFromMonitor(cr, monitor); err != nil {
		logger.Error(err, "Failed to update status from monitor")
		return managed.ExternalObservation{}, err
	}

	// Set conditions based on resource state
	c.setObserveConditions(cr, resourceUpToDate, monitorID, logger)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: resourceUpToDate,
	}, nil
}
