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
	"encoding/json"
	"reflect"

	"github.com/crossplane/crossplane-runtime/pkg/feature"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/statemetrics"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	meta "github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/provider-komodor/apis/komodor/v1alpha1"
	apisv1alpha1 "github.com/crossplane/provider-komodor/apis/v1alpha1"
	komodorclient "github.com/crossplane/provider-komodor/internal/clients/komodor"
	"github.com/crossplane/provider-komodor/internal/features"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	errNotRealtimeMonitor = "managed resource is not a RealtimeMonitor custom resource"
	errTrackPCUsage       = "cannot track ProviderConfig usage"
	errGetPC              = "cannot get ProviderConfig"
	errGetCreds           = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// Define KomodorClient interface for testability
type KomodorClient interface {
	GetMonitor(ctx context.Context, id string) (*komodorclient.Monitor, error)
	CreateMonitor(ctx context.Context, monitor *komodorclient.Monitor) (*komodorclient.Monitor, error)
	UpdateMonitor(ctx context.Context, id string, monitor *komodorclient.Monitor) (*komodorclient.Monitor, error)
	DeleteMonitor(ctx context.Context, id string) error
}

// A NoOpService does nothing.
type NoOpService struct{}

var (
	newKomodorClient = func(apiKey []byte) (interface{}, error) {
		return komodorclient.NewClient(string(apiKey)), nil
	}
)

// Setup adds a controller that reconciles RealtimeMonitor managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RealtimeMonitorGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newKomodorClient}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
		managed.WithManagementPolicies(),
	}

	if o.Features.Enabled(feature.EnableAlphaChangeLogs) {
		opts = append(opts, managed.WithChangeLogger(o.ChangeLogOptions.ChangeLogger))
	}

	if o.MetricOptions != nil {
		opts = append(opts, managed.WithMetricRecorder(o.MetricOptions.MRMetrics))
	}

	if o.MetricOptions != nil && o.MetricOptions.MRStateMetrics != nil {
		stateMetricsRecorder := statemetrics.NewMRStateRecorder(
			mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.RealtimeMonitorList{}, o.MetricOptions.PollStateMetricInterval,
		)
		if err := mgr.Add(stateMetricsRecorder); err != nil {
			return errors.Wrap(err, "cannot register MR state metrics recorder for kind v1alpha1.RealtimeMonitorList")
		}
	}

	r := managed.NewReconciler(mgr, resource.ManagedKind(v1alpha1.RealtimeMonitorGroupVersionKind), opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.RealtimeMonitor{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(creds []byte) (interface{}, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return nil, errors.New(errNotRealtimeMonitor)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(data)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	client, ok := svc.(*komodorclient.Client)
	if !ok {
		return nil, errors.New("failed to cast to Komodor client")
	}

	return &external{client: client}, nil
}

// external implements managed.ExternalClient using the Komodor client.
type external struct {
	client KomodorClient
}

// Helper: Unmarshal apiextensionsv1.JSON slice to []map[string]interface{}
func unmarshalSensors(jsons []apiextensionsv1.JSON) ([]map[string]interface{}, error) {
	sensors := make([]map[string]interface{}, 0, len(jsons))
	for _, s := range jsons {
		var m map[string]interface{}
		if err := json.Unmarshal(s.Raw, &m); err != nil {
			return nil, err
		}
		sensors = append(sensors, m)
	}
	return sensors, nil
}

// Helper: Unmarshal apiextensionsv1.JSON to map[string]interface{}
func unmarshalMap(j apiextensionsv1.JSON) (map[string]interface{}, error) {
	if j.Raw == nil {
		return nil, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(j.Raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Helper: Marshal []map[string]interface{} to []apiextensionsv1.JSON
func marshalSensors(sensors []map[string]interface{}) ([]apiextensionsv1.JSON, error) {
	jsons := make([]apiextensionsv1.JSON, 0, len(sensors))
	for _, s := range sensors {
		b, err := json.Marshal(s)
		if err != nil {
			return nil, err
		}
		jsons = append(jsons, apiextensionsv1.JSON{Raw: b})
	}
	return jsons, nil
}

// Helper: Marshal map[string]interface{} to apiextensionsv1.JSON
func marshalMap(m map[string]interface{}) (apiextensionsv1.JSON, error) {
	if m == nil {
		return apiextensionsv1.JSON{}, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return apiextensionsv1.JSON{}, err
	}
	return apiextensionsv1.JSON{Raw: b}, nil
}

// Helper: Update status fields from a Monitor
func updateStatusFromMonitor(cr *v1alpha1.RealtimeMonitor, m *komodorclient.Monitor) error {
	cr.Status.AtProvider.ID = m.ID
	cr.Status.AtProvider.Name = m.Name
	cr.Status.AtProvider.Active = m.Active
	cr.Status.AtProvider.Type = m.Type
	jsons, err := marshalSensors(m.Sensors)
	if err != nil {
		return errors.Wrap(err, "failed to marshal sensors for status")
	}
	cr.Status.AtProvider.Sensors = jsons
	if m.Sinks != nil {
		j, err := marshalMap(m.Sinks)
		if err != nil {
			return errors.Wrap(err, "failed to marshal sinks for status")
		}
		cr.Status.AtProvider.Sinks = j
	}
	if m.Variables != nil {
		j, err := marshalMap(m.Variables)
		if err != nil {
			return errors.Wrap(err, "failed to marshal variables for status")
		}
		cr.Status.AtProvider.Variables = j
	}
	cr.Status.AtProvider.SinksOptions = m.SinksOptions
	cr.Status.AtProvider.CreatedAt = m.CreatedAt
	cr.Status.AtProvider.UpdatedAt = m.UpdatedAt
	cr.Status.AtProvider.IsDeleted = m.IsDeleted
	return nil
}

// Helper: Compare spec and monitor for up-to-date status
func isMonitorUpToDate(spec *v1alpha1.RealtimeMonitorParameters, monitor *komodorclient.Monitor, specSensors []map[string]interface{}, specSinks, specVariables map[string]interface{}) bool {
	return spec.Name == monitor.Name &&
		reflect.DeepEqual(specSensors, monitor.Sensors) &&
		reflect.DeepEqual(specSinks, monitor.Sinks) &&
		spec.Active == monitor.Active &&
		spec.Type == monitor.Type &&
		reflect.DeepEqual(specVariables, monitor.Variables) &&
		reflect.DeepEqual(spec.SinksOptions, monitor.SinksOptions)
}

// Helper: Handle error from Komodor GetMonitor
func handleGetMonitorError(cr *v1alpha1.RealtimeMonitor, extName string, err error) (managed.ExternalObservation, error) {
	if komodorclient.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot get monitor from Komodor")))
	return managed.ExternalObservation{}, errors.Wrapf(err, "failed to get monitor %q from Komodor", extName)
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRealtimeMonitor)
	}

	// If no external name, resource does not exist
	extName := meta.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	monitor, err := c.client.GetMonitor(ctx, extName)
	if err != nil {
		return handleGetMonitorError(cr, extName, err)
	}

	specSensors, err := unmarshalSensors(cr.Spec.ForProvider.Sensors)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to unmarshal spec sensors")
	}
	specSinks, err := unmarshalMap(cr.Spec.ForProvider.Sinks)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to unmarshal spec sinks")
	}
	specVariables, err := unmarshalMap(cr.Spec.ForProvider.Variables)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to unmarshal spec variables")
	}

	resourceUpToDate := isMonitorUpToDate(&cr.Spec.ForProvider, monitor, specSensors, specSinks, specVariables)

	if err := updateStatusFromMonitor(cr, monitor); err != nil {
		return managed.ExternalObservation{}, err
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: resourceUpToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRealtimeMonitor)
	}

	sensors, err := unmarshalSensors(cr.Spec.ForProvider.Sensors)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to unmarshal spec sensors")
	}
	sinks, err := unmarshalMap(cr.Spec.ForProvider.Sinks)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to unmarshal spec sinks")
	}
	variables, err := unmarshalMap(cr.Spec.ForProvider.Variables)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to unmarshal spec variables")
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

	created, err := c.client.CreateMonitor(ctx, monitor)
	if err != nil {
		cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot create monitor in Komodor")))
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create monitor in Komodor")
	}

	meta.SetExternalName(cr, created.ID)

	if err := updateStatusFromMonitor(cr, created); err != nil {
		return managed.ExternalCreation{}, err
	}

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRealtimeMonitor)
	}

	extName := meta.GetExternalName(cr)
	if extName == "" {
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

	updated, err := c.client.UpdateMonitor(ctx, extName, monitor)
	if err != nil {
		cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot update monitor in Komodor")))
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update monitor in Komodor")
	}

	if err := updateStatusFromMonitor(cr, updated); err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.RealtimeMonitor)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRealtimeMonitor)
	}

	extName := meta.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalDelete{}, errors.New("external name (monitor ID) is not set")
	}

	if err := c.client.DeleteMonitor(ctx, extName); err != nil {
		cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot delete monitor in Komodor")))
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot delete monitor in Komodor")
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}
