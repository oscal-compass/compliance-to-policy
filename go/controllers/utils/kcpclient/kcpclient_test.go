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

package kcpclient

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/oscal-compass/compliance-to-policy/go/controllers/testsetting"
	"github.com/oscal-compass/compliance-to-policy/go/pkg"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var testSetting *testsetting.KcpTestSetting
var testdata = pkg.PathFromPkgDirectory("../controllers/utils/kcpclient/testdata")
var serverReadinessTimeout time.Duration = time.Second * 60
var serverReadinessInterval time.Duration = time.Second * 10
var timeout time.Duration = time.Second * 10
var interval time.Duration = time.Second * 1

func TestKcpClient(t *testing.T) {
	if _, ok := os.LookupEnv("DO_NOT_SKIP_ANY_TEST"); !ok {
		t.Skip("Skipping testing")
	}
	RegisterFailHandler(Fail)

	RunSpecs(t, "KcpClient Test Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error
	testSetting, err = testsetting.KcpSetup(time.Second * 30)
	Expect(err).NotTo(HaveOccurred())
	Expect(testSetting).NotTo(BeNil())

	Eventually(func() bool {
		return isKcpReady(testSetting.Cfg, testSetting.Ctx)
	}, serverReadinessTimeout, serverReadinessInterval).Should(BeTrue())

	// Create workspaces
	kcpClient, err := NewKcpClient(*testSetting.Cfg, "root")
	Expect(err).NotTo(HaveOccurred())

	wsDyClient, err := kcpClient.GetDyClient("tenancy.kcp.io", "Workspace", "v1alpha1")
	Expect(err).NotTo(HaveOccurred())

	workspaceManifests := []string{testdata + "/ws.test1.yaml", testdata + "/ws.test2.yaml"}
	for _, workspaceManifest := range workspaceManifests {

		var unst *unstructured.Unstructured
		err = pkg.LoadYamlFileToObject(workspaceManifest, &unst)
		Expect(err).NotTo(HaveOccurred())

		testSetting.CreateWorkspace(timeout, interval, wsDyClient, unst)
	}
})

var _ = AfterSuite(func() {
	testSetting.Cancel()
	if err := testSetting.TerminateKcp(); err != nil {
		Expect(err).NotTo(HaveOccurred())
	}
})

var _ = Describe("Test KcpClient", func() {

	Context("When accessing workspace", func() {
		It("should list namespaces in the given workspace", func() {
			// time.Sleep(time.Second * 5) // just workaround (not sure why getting namespace failed even if workspace is ready)
			// cfg, _ = getKubeconfig()
			kcpClient1, err := NewKcpClient(*testSetting.Cfg, "root:test1")
			Expect(err).NotTo(HaveOccurred())
			var namespaceList corev1.NamespaceList
			err = kcpClient1.K8sClient.List(testSetting.Ctx, &namespaceList, &client.ListOptions{})
			Expect(err).NotTo(HaveOccurred())

			kcpClient2, err := NewKcpClient(*testSetting.Cfg, "root:test2")
			Expect(err).NotTo(HaveOccurred())
			err = kcpClient2.K8sClient.List(testSetting.Ctx, &namespaceList, &client.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

})

func isKcpReady(cfg *rest.Config, ctx context.Context) bool { // TODO: Externalize this to common test modules
	var err error
	kcpClient, err := NewKcpClient(*cfg, "root")
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
