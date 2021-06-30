// Copyright 2020-2021 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	apiutils "github.com/clastix/capsule/api/utils"
	capsulev1alpha1 "github.com/clastix/capsule/api/v1alpha1"
	capsulewebhook "github.com/clastix/capsule/pkg/webhook"
	"github.com/clastix/capsule/pkg/webhook/utils"
)

type containerRegistryHandler struct {
}

func ContainerRegistry() capsulewebhook.Handler {
	return &containerRegistryHandler{}
}

func (h *containerRegistryHandler) OnCreate(c client.Client, decoder *admission.Decoder, recorder record.EventRecorder) capsulewebhook.Func {
	return func(ctx context.Context, req admission.Request) *admission.Response {
		pod := &corev1.Pod{}
		if err := decoder.Decode(req, pod); err != nil {
			return utils.ErroredResponse(err)
		}

		tntList := &capsulev1alpha1.TenantList{}
		if err := c.List(ctx, tntList, client.MatchingFieldsSelector{
			Selector: fields.OneTermEqualSelector(".status.namespaces", pod.Namespace),
		}); err != nil {
			return utils.ErroredResponse(err)
		}

		if len(tntList.Items) == 0 {
			return nil
		}

		tnt := tntList.Items[0]

		if tnt.Spec.ContainerRegistries != nil {
			var valid, matched bool

			for _, container := range pod.Spec.Containers {
				registry := apiutils.NewRegistry(container.Image)

				valid = tnt.Spec.ContainerRegistries.ExactMatch(registry.Registry())

				matched = tnt.Spec.ContainerRegistries.RegexMatch(registry.Registry())

				if !valid && !matched {
					recorder.Eventf(&tnt, corev1.EventTypeWarning, "ForbiddenContainerRegistry", "Pod %s/%s is using a forbidden registry %s is forbidden for the current Tenant", req.Namespace, req.Name, registry.Registry())

					response := admission.Denied(NewContainerRegistryForbidden(container.Image, *tnt.Spec.ContainerRegistries).Error())

					return &response
				}
			}
		}

		return nil
	}
}

func (h *containerRegistryHandler) OnDelete(client.Client, *admission.Decoder, record.EventRecorder) capsulewebhook.Func {
	return func(ctx context.Context, req admission.Request) *admission.Response {
		return nil
	}
}

func (h *containerRegistryHandler) OnUpdate(client.Client, *admission.Decoder, record.EventRecorder) capsulewebhook.Func {
	return func(ctx context.Context, req admission.Request) *admission.Response {
		return nil
	}
}
