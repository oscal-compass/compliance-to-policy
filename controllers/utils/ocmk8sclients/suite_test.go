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

package ocmk8sclients

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	compliancetopolicycontrollerv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/pkg"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var dyClient dynamic.Interface
var ocmK8ResourceInterfaceSet OcmK8ResourceInterfaceSetType
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc
var timeout = time.Second * 10
var interval = time.Second * 1
var sampleNamespace = "sample"
var testdataDir = pkg.PathFromPkgDirectory("../controllers/utils/ocmk8sclients/testdata")

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{pkg.PathFromPkgDirectory("../config/crd/bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = compliancetopolicycontrollerv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	dyClient, err = dynamic.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(dyClient).NotTo(BeNil())

	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(cfg)

	apixClientSet, err := apix.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(apixClientSet).NotTo(BeNil())

	// add CRDs
	crdDir := pkg.PathFromPkgDirectory("../config/ocm")
	files, err := os.ReadDir(crdDir)
	if err != nil {
		panic(err)
	}

	if err := k8sClient.Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: sampleNamespace,
		},
	}, &client.CreateOptions{}); err != nil {
		panic(err)
	}

	for _, file := range files {
		path := crdDir + "/" + file.Name()
		crd := apixv1.CustomResourceDefinition{}
		if err := pkg.LoadYamlFileToK8sTypedObject(path, &crd); err != nil {
			panic(err)
		}
		_crd, err := apixClientSet.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), &crd, v1.CreateOptions{})
		if err != nil {
			panic(err)
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

	Eventually(func() bool {
		ocmK8ResourceInterfaceSet, err = NewOcmK8sClientSet(discoveryClient, dyClient)
		if err != nil {
			GinkgoWriter.Printf("failed to initialize ocmK8sClientSet: %v", err)
			return false
		}
		return true
	}, timeout, interval).Should(BeTrue())

})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
