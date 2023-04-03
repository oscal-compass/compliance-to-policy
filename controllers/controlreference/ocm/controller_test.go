/*
Copyright 2023 IBM Corporation

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

package ocm

import (
	"context"

	ctrlv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/utils/ocmk8sclients"
	"github.com/IBM/compliance-to-policy/pkg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"os"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes/scheme"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/IBM/compliance-to-policy/controllers/testsetting"
)

var testSetting *testsetting.TestSetting
var namespace = "default"
var testdataDir = pkg.PathFromPkgDirectory("../controllers/testdata")

func TestControlReferenceOcmController(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "ControlReference Controller Test")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error
	testSetting, err = testsetting.Setup(time.Second*10, time.Second*1)
	Expect(err).NotTo(HaveOccurred())

	k8sManager, err := ctrl.NewManager(testSetting.Cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	tempDir := pkg.PathFromPkgDirectory("../controllers/controlreference/ocm/_test")
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		panic(err)
	}

	err = (&ControlReferenceReconciler{
		Client:                    k8sManager.GetClient(),
		Scheme:                    k8sManager.GetScheme(),
		TempDir:                   tempDir,
		OcmK8ResourceInterfaceSet: testSetting.OcmK8ResourceInterfaceSet,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(testSetting.Ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	testSetting.Cancel()
	By("tearing down the test environment")
	err := testSetting.TestEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Test Control Reference Controller", func() {

	controlReference := ctrlv1alpha1.ControlReference{}
	err := pkg.LoadYamlFileToK8sTypedObject(testdataDir+"/controlreference.yaml", &controlReference)
	Expect(err).NotTo(HaveOccurred())

	Context("When creating ControlReference", func() {
		It("should create the object", func() {
			err := testSetting.K8sClient.Create(context.TODO(), &controlReference, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("should list the created object", func() {
			controlReferenceList := ctrlv1alpha1.ControlReferenceList{}
			err := testSetting.K8sClient.List(context.TODO(), &controlReferenceList, &client.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(controlReferenceList.Items)).To(Equal(1))
		})
		It("should generate Policy, PlacementRule, PlacementBinding", func() {
			Eventually(func() int {
				client := ocmk8sclients.NewPolicyClient(testSetting.OcmK8ResourceInterfaceSet.Policy)
				typedList, err := client.List(namespace)
				if err != nil {
					GinkgoWriter.Printf("failed to get Policy: %v", err)
					return 0
				}
				return len(typedList)
			}, testSetting.Timeout, testSetting.Interval).Should(Equal(2))
			Eventually(func() int {
				client := ocmk8sclients.NewPlacementBindingClient(testSetting.OcmK8ResourceInterfaceSet.PlacementBinding)
				typedList, err := client.List(namespace)
				if err != nil {
					GinkgoWriter.Printf("failed to get PlacementBinding: %v", err)
					return 0
				}
				return len(typedList)
			}, testSetting.Timeout, testSetting.Interval).Should(Equal(2))
			Eventually(func() int {
				client := ocmk8sclients.NewPlacementRuleClient(testSetting.OcmK8ResourceInterfaceSet.PlacementRule)
				typedList, err := client.List(namespace)
				if err != nil {
					GinkgoWriter.Printf("failed to get PlacementRule: %v", err)
					return 0
				}
				return len(typedList)
			}, testSetting.Timeout, testSetting.Interval).Should(Equal(2))
		})
	})
})
