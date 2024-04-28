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

package testsetting

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	compliancetopolicycontrollerv1alpha1 "github.com/oscal-compass/compliance-to-policy/go/api/v1alpha1"
	"github.com/oscal-compass/compliance-to-policy/go/controllers/utils/ocmk8sclients"
	"github.com/oscal-compass/compliance-to-policy/go/pkg"

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
)

type TestSetting struct {
	Cfg                       *rest.Config
	K8sClient                 client.Client
	DyClient                  dynamic.Interface
	OcmK8ResourceInterfaceSet ocmk8sclients.OcmK8ResourceInterfaceSetType
	TestEnv                   *envtest.Environment
	Ctx                       context.Context
	Cancel                    context.CancelFunc
	Timeout                   time.Duration
	Interval                  time.Duration
	SampleNamespace           string
}

func Setup(timout time.Duration, interval time.Duration) (*TestSetting, error) {

	t := TestSetting{}
	t.Timeout = timout
	t.Interval = interval
	t.Ctx, t.Cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	t.TestEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{pkg.PathFromPkgDirectory("../config/crd/bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	t.Cfg, err = t.TestEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(t.Cfg).NotTo(BeNil())

	err = compliancetopolicycontrollerv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	t.K8sClient, err = client.New(t.Cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(t.K8sClient).NotTo(BeNil())

	t.DyClient, err = dynamic.NewForConfig(t.Cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(t.DyClient).NotTo(BeNil())

	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(t.Cfg)

	apixClientSet, err := apix.NewForConfig(t.Cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(apixClientSet).NotTo(BeNil())

	// add CRDs
	crdDir := pkg.PathFromPkgDirectory("../config/ocm")
	files, err := os.ReadDir(crdDir)
	if err != nil {
		return nil, err
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
		}, t.Timeout, t.Interval).Should(BeTrue())
	}

	Eventually(func() bool {
		t.OcmK8ResourceInterfaceSet, err = ocmk8sclients.NewOcmK8sClientSet(discoveryClient, t.DyClient)
		if err != nil {
			GinkgoWriter.Printf("failed to initialize ocmK8sClientSet: %v", err)
			return false
		}
		return true
	}, t.Timeout, t.Interval).Should(BeTrue())

	return &t, nil
}

func (t *TestSetting) CreateNamespace(namespace string) {
	if err := t.K8sClient.Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: namespace,
		},
	}, &client.CreateOptions{}); err != nil {
		panic(err)
	}
}

func (t *TestSetting) CreateNamespacedObj(namespace string, obj client.Object) {
	obj.SetNamespace(namespace)
	if err := t.K8sClient.Create(context.TODO(), obj, &client.CreateOptions{}); err != nil {
		panic(err)
	}
}
