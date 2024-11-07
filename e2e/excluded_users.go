//go:build e2e

// Copyright 2020-2023 Project Capsule Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

var _ = Describe("creating a Namespace with a User from ExcludedUsers in Capsule Configuration", func() {
	t1 := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "awesome",
		},
		Spec: capsulev1beta2.TenantSpec{
			Owners: capsulev1beta2.OwnerListSpec{
				{
					Name: "system:serviceaccount:gitops-namespace:excluded-service-account",
					Kind: "ServiceAccount",
				},
			},
		},
	}
	ns := NewNamespace("excluded-namespace")

	JustBeforeEach(func() {
		EventuallyCreation(func() error {
			t1.ResourceVersion = ""
			return k8sClient.Create(context.TODO(), t1)
		}).Should(Succeed())
	})
	JustAfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), ns)).Should(Succeed())
		Expect(k8sClient.Delete(context.TODO(), t1)).Should(Succeed())
		ModifyCapsuleConfigurationOpts(func(configuration *capsulev1beta2.CapsuleConfiguration) {
			configuration.Spec.ExcludedUsers = []string{}
		})
	})

	It("should create the Namespace and NOT be available in the Tenant List", func() {
		ModifyCapsuleConfigurationOpts(func(configuration *capsulev1beta2.CapsuleConfiguration) {
			configuration.Spec.ExcludedUsers = []string{t1.Spec.Owners[0].Name}
		})

		NamespaceCreation(ns, t1.Spec.Owners[0], defaultTimeoutInterval).Should(Succeed())
		TenantNamespaceList(t1, defaultTimeoutInterval).ShouldNot(ContainElement(ns.GetName()))
	})
})
