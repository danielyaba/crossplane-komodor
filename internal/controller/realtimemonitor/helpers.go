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
	"encoding/json"
	"reflect"
	"regexp"

	"github.com/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/crossplane/provider-komodor/apis/komodor/v1alpha1"
	komodorclient "github.com/crossplane/provider-komodor/internal/clients/komodor"
)

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

// isValidUUID checks if a string is a valid UUID format
func isValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(uuid)
}

// Helper struct for spec data
type specData struct {
	sensors   []map[string]interface{}
	sinks     map[string]interface{}
	variables map[string]interface{}
}

// Helper: Unmarshal spec data
func unmarshalSpecData(cr *v1alpha1.RealtimeMonitor) (*specData, error) {
	specSensors, err := unmarshalSensors(cr.Spec.ForProvider.Sensors)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal spec sensors")
	}

	specSinks, err := unmarshalMap(cr.Spec.ForProvider.Sinks)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal spec sinks")
	}

	specVariables, err := unmarshalMap(cr.Spec.ForProvider.Variables)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal spec variables")
	}

	return &specData{
		sensors:   specSensors,
		sinks:     specSinks,
		variables: specVariables,
	}, nil
}
