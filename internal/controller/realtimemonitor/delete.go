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
	"sigs.k8s.io/controller-runtime/pkg/log"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/provider-komodor/apis/komodor/v1alpha1"
)

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	logger := log.FromContext(ctx)

	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRealtimeMonitor)
	}

	logger.Info("Deleting RealtimeMonitor",
		"name", cr.Name,
		"namespace", cr.Namespace)

	extName := meta.GetExternalName(cr)
	if extName == "" {
		logger.Error(errors.New("external name not set"), "Cannot delete monitor without external name")
		return managed.ExternalDelete{}, errors.New("external name (monitor ID) is not set")
	}

	logger.Info("Sending delete request to Komodor", "monitorID", extName)

	if err := c.client.DeleteMonitor(ctx, extName); err != nil {
		logger.Error(err, "Failed to delete monitor in Komodor", "monitorID", extName)
		cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot delete monitor in Komodor")))
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot delete monitor in Komodor")
	}

	logger.Info("Successfully deleted monitor in Komodor", "monitorID", extName)
	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}
