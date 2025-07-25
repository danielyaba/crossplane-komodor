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

	meta "github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/provider-komodor/apis/komodor/v1alpha1"
	komodorclient "github.com/crossplane/provider-komodor/internal/clients/komodor"
)

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRealtimeMonitor)
	}

	// Get the monitor ID from external-name annotation
	monitorID := meta.GetExternalName(cr)
	if monitorID == "" {
		return managed.ExternalUpdate{}, errors.New("external name (monitor ID) is not set")
	}

	sensors, err := unmarshalSensors(cr.Spec.ForProvider.Sensors)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to unmarshal spec sensors")
	}
	sinks, err := unmarshalMap(cr.Spec.ForProvider.Sinks)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to unmarshal spec sinks")
	}
	variables, err := unmarshalMap(cr.Spec.ForProvider.Variables)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to unmarshal spec variables")
	}

	monitor := &komodorclient.Monitor{
		Name:         cr.Spec.ForProvider.Name,
		Sensors:      sensors,
		Sinks:        sinks,
		Active:       cr.Spec.ForProvider.Active,
		Type:         cr.Spec.ForProvider.Type,
		Variables:    variables,
		SinksOptions: cr.Spec.ForProvider.SinksOptions,
	}

	updated, err := c.client.UpdateMonitor(ctx, monitorID, monitor)
	if err != nil {
		cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot update monitor in Komodor")))
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update monitor in Komodor")
	}

	if err := updateStatusFromMonitor(cr, updated); err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}
