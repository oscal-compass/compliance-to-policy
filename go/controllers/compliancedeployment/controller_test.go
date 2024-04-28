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

package compliancedeployment

import (
	"context"

	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrlv1alpha1 "github.com/oscal-compass/compliance-to-policy/go/api/v1alpha1"
	"github.com/oscal-compass/compliance-to-policy/go/controllers/utils/ocmk8sclients"
	"github.com/oscal-compass/compliance-to-policy/go/pkg"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes/scheme"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	ctrlrefocm "github.com/oscal-compass/compliance-to-policy/go/controllers/controlreference/ocm"
	"github.com/oscal-compass/compliance-to-policy/go/controllers/testsetting"
)

var testSetting *testsetting.TestSetting
var sampleNamespace = "sample"
var complianceDeploymentTestNamespace = "test-cd"
var testdataDir = pkg.PathFromPkgDirectory("../controllers/testdata")

func TestComplianceDeploymentController(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "ComplianceDeployment Controller Test")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error
	testSetting, err = testsetting.Setup(time.Second*10, time.Second*1)
	Expect(err).ToNot(HaveOccurred())
	testSetting.CreateNamespace(sampleNamespace)
	testSetting.CreateNamespace(complianceDeploymentTestNamespace)

	k8sManager, err := ctrl.NewManager(testSetting.Cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	tempDir := pkg.PathFromPkgDirectory("../controllers/compliancedeployment/_test")
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		panic(err)
	}

	err = (&ComplianceDeploymentReconciler{
		Client:  k8sManager.GetClient(),
		Scheme:  k8sManager.GetScheme(),
		TempDir: tempDir,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&ctrlrefocm.ControlReferenceReconciler{
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

var _ = Describe("Test ComplianceDeployment Controller", func() {

	complianceDeployment := ctrlv1alpha1.ComplianceDeployment{}
	err := pkg.LoadYamlFileToK8sTypedObject(testdataDir+"/compliancedeployment.yaml", &complianceDeployment)
	complianceDeployment.SetNamespace(complianceDeploymentTestNamespace)
	Expect(err).NotTo(HaveOccurred())
	numberOfGeneratedPolicy := 85

	Context("When creating ComplianceDeployment", func() {
		It("should create the object", func() {
			err := testSetting.K8sClient.Create(context.TODO(), &complianceDeployment, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("should list the created object", func() {
			Eventually(func() int {
				complianceDeploymentList := ctrlv1alpha1.ComplianceDeploymentList{}
				err := testSetting.K8sClient.List(context.TODO(), &complianceDeploymentList, &client.ListOptions{})
				if err != nil {
					return 0
				}
				return len(complianceDeploymentList.Items)
			}, testSetting.Timeout, testSetting.Interval).Should(Equal(1))
		})
		It("should create secondary object", func() {
			var controlReference ctrlv1alpha1.ControlReference
			Eventually(func() bool {
				namespacedName := types.NamespacedName{
					Namespace: complianceDeploymentTestNamespace,
					Name:      complianceDeployment.Name,
				}
				err := testSetting.K8sClient.Get(context.TODO(), namespacedName, &controlReference)
				return err == nil
			}, testSetting.Timeout, testSetting.Interval).Should(Equal(true))
			Expect(err).NotTo(HaveOccurred())
			Expect(controlReference).NotTo(BeNil())
		})
		It("should generate Policy, PlacementRule, PlacementBinding", func() {
			targetNamespace := complianceDeployment.Spec.Target.Namespace
			Eventually(func() int {
				client := ocmk8sclients.NewPolicyClient(testSetting.OcmK8ResourceInterfaceSet.Policy)
				typedList, err := client.List(targetNamespace)
				if err != nil {
					GinkgoWriter.Printf("failed to get Policy: %v", err)
					return 0
				}
				return len(typedList)
			}, testSetting.Timeout, testSetting.Interval).Should(Equal(numberOfGeneratedPolicy))
			Eventually(func() int {
				client := ocmk8sclients.NewPlacementBindingClient(testSetting.OcmK8ResourceInterfaceSet.PlacementBinding)
				typedList, err := client.List(targetNamespace)
				if err != nil {
					GinkgoWriter.Printf("failed to get PlacementBinding: %v", err)
					return 0
				}
				return len(typedList)
			}, testSetting.Timeout, testSetting.Interval).Should(Equal(numberOfGeneratedPolicy))
			Eventually(func() int {
				client := ocmk8sclients.NewPlacementRuleClient(testSetting.OcmK8ResourceInterfaceSet.PlacementRule)
				typedList, err := client.List(targetNamespace)
				if err != nil {
					GinkgoWriter.Printf("failed to get PlacementRule: %v", err)
					return 0
				}
				return len(typedList)
			}, testSetting.Timeout, testSetting.Interval).Should(Equal(numberOfGeneratedPolicy))
		})
	})
})
