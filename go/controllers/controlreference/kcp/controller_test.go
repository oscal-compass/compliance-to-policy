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

package kcp

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrlv1alpha1 "github.com/oscal-compass/compliance-to-policy/go/api/v1alpha1"
	"github.com/oscal-compass/compliance-to-policy/go/controllers/testsetting"
	"github.com/oscal-compass/compliance-to-policy/go/controllers/utils/kcpclient"
	"github.com/oscal-compass/compliance-to-policy/go/pkg"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var t *testsetting.KcpTestSetting
var testdata = pkg.PathFromPkgDirectory("../controllers/testdata")
var serverReadinessTimeout time.Duration = time.Second * 60
var serverReadinessInterval time.Duration = time.Second * 10
var timeout time.Duration = time.Second * 10
var interval time.Duration = time.Second * 1

func TestControlReferenceKcpController(t *testing.T) {
	if _, ok := os.LookupEnv("DO_NOT_SKIP_ANY_TEST"); !ok {
		t.Skip("Skipping testing")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "ControlReference Controller for KCP Test")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error
	t, err = testsetting.KcpSetup(time.Second * 30)
	Expect(err).NotTo(HaveOccurred())
	Expect(t).NotTo(BeNil())

	Eventually(func() bool {
		return isKcpReady(t.Cfg, t.Ctx)
	}, serverReadinessTimeout, serverReadinessInterval).Should(BeTrue())

	// Create workspaces
	kcpClient, err := kcpclient.NewKcpClient(*t.Cfg, "root")
	Expect(err).NotTo(HaveOccurred())
	wsDyClient, err := kcpClient.GetDyClient("tenancy.kcp.io", "Workspace", "v1alpha1")
	Expect(err).NotTo(HaveOccurred())

	workspaceManifests := []string{testdata + "/ws.test1.yaml", testdata + "/ws.test2.yaml"}
	for _, workspaceManifest := range workspaceManifests {

		var unst *unstructured.Unstructured
		err = pkg.LoadYamlFileToObject(workspaceManifest, &unst)
		Expect(err).NotTo(HaveOccurred())

		t.CreateWorkspace(timeout, interval, wsDyClient, unst)
	}

	apixClientSet, err := apix.NewForConfig(t.Cfg)
	Expect(err).NotTo(HaveOccurred())

	// add CRDs
	crdDir := pkg.PathFromPkgDirectory("../config/crd/bases")
	files, err := os.ReadDir(crdDir)
	Expect(err).NotTo(HaveOccurred())

	for _, file := range files {
		if file.Name() == "_.yaml" {
			continue
		}
		path := crdDir + "/" + file.Name()
		crd := apixv1.CustomResourceDefinition{}
		if err := pkg.LoadYamlFileToK8sTypedObject(path, &crd); err != nil {
			panic(err)
		}
		_crd, err := apixClientSet.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), &crd, v1.CreateOptions{})
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				continue
			} else {
				panic(err)
			}
		}
		Eventually(func() bool {
			_, err := apixClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), _crd.Name, v1.GetOptions{})
			if err != nil {
				GinkgoWriter.Printf("failed to get CRD %s: %v", _crd.Name, err)
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
	}

	k8sManager, err := ctrl.NewManager(t.Cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	tempDir := pkg.PathFromPkgDirectory("../controllers/controlreference/kcp/_test")
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		panic(err)
	}

	err = (&ControlReferenceKcpReconciler{
		Client:  k8sManager.GetClient(),
		Scheme:  k8sManager.GetScheme(),
		TempDir: tempDir,
		Cfg:     t.Cfg,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(t.Ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	t.Cancel()
	if err := t.TerminateKcp(); err != nil {
		Expect(err).NotTo(HaveOccurred())
	}
})

var _ = Describe("Test Control Reference Controller for KCP", func() {

	cr := ctrlv1alpha1.ControlReferenceKcp{}
	err := pkg.LoadYamlFileToK8sTypedObject(testdata+"/controlreferencekcp.yaml", &cr)
	Expect(err).NotTo(HaveOccurred())

	Context("When ControlReference is created", func() {
		It("should create the object", func() {
			err := t.K8sClient.Create(t.Ctx, &cr, &client.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("should list the created object", func() {
			crList := ctrlv1alpha1.ControlReferenceKcpList{}
			err := t.K8sClient.List(context.TODO(), &crList, &client.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(crList.Items)).To(Equal(1))
		})
		It("should generate ConfigPolicy in each workspace", func() {
			workspaces := []string{"root:ws1", "root:ws2"}
			for _, workspace := range workspaces {
				kcpClient, err := kcpclient.NewKcpClient(*t.Cfg, workspace)
				Expect(err).NotTo(HaveOccurred())
				dyClient, err := kcpClient.GetDyClient("policy.open-cluster-management.io", "ConfigurationPolicy", "v1")
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() int {
					configPolicyUnstList, err := dyClient.Namespace("default").List(t.Ctx, v1.ListOptions{})
					if err != nil {
						return 0
					}
					return len(configPolicyUnstList.Items)
				}, time.Second*60, time.Second*10).Should(Equal(5))
			}
		})
	})

})

func isKcpReady(cfg *rest.Config, ctx context.Context) bool { // TODO: Externalize this to common test modules
	var err error
	kcpClient, err := kcpclient.NewKcpClient(*cfg, "root")
	if err != nil {
		return false
	}
	wsDyClient, err := kcpClient.GetDyClient("tenancy.kcp.io", "Workspace", "v1alpha1")
	if err != nil {
		return false
	}
	unstWsList, err := wsDyClient.List(ctx, v1.ListOptions{})
	if err != nil {
		return false
	}
	return len(unstWsList.Items) > 0
}
