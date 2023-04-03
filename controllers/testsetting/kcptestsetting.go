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
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	compliancetopolicycontrollerv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/utils/ocmk8sclients"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var KCP_ROOTDIR = os.Getenv("KCP_ROOTDIR")
var KCP_BIN = os.Getenv("KCP_BIN")
var KCP_KUBECONFIG = os.Getenv("KCP_KUBECONFIG")
var KCP_ISRUNNING = os.Getenv("KCP_ISRUNNING")

type KcpTestSetting struct {
	Cfg                       *rest.Config
	K8sClient                 client.Client
	OcmK8ResourceInterfaceSet ocmk8sclients.OcmK8ResourceInterfaceSetType
	Ctx                       context.Context
	Cancel                    context.CancelFunc
	kcpPid                    int
}

func KcpSetup(timeout time.Duration) (*KcpTestSetting, error) {

	By("bootstrapping test environment")

	t := KcpTestSetting{}
	t.Ctx, t.Cancel = context.WithCancel(context.TODO())

	var err error
	if KCP_ISRUNNING == "true" {
		t.Cfg, err = getKubeconfig()
	} else {
		t.Cfg, err = t.runKcp(timeout, time.Second*10)
	}
	Expect(err).NotTo(HaveOccurred())

	err = compliancetopolicycontrollerv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	t.K8sClient, err = client.New(t.Cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(t.K8sClient).NotTo(BeNil())

	return &t, err
}

func (t *KcpTestSetting) CreateWorkspace(timeout time.Duration, interval time.Duration, wsDyClient dynamic.NamespaceableResourceInterface, unst *unstructured.Unstructured) {
	unstWs, err := wsDyClient.Get(t.Ctx, unst.GetName(), v1.GetOptions{})
	if err == nil && unstWs != nil {
		return
	}
	unst, err = wsDyClient.Create(t.Ctx, unst, v1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())
	Eventually(func() bool {
		unst, err := wsDyClient.Get(t.Ctx, unst.GetName(), v1.GetOptions{})
		if err != nil || unst == nil {
			return false
		}
		phase, ok, err := unstructured.NestedString(unst.Object, "status", "phase")
		if !ok || err != nil {
			return false
		}
		return phase == "Ready"
	}, timeout, interval).Should(BeTrue())
}

func getKubeconfigFromFile(path string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (t *KcpTestSetting) runKcp(timeout time.Duration, interval time.Duration) (*rest.Config, error) {
	if err := os.RemoveAll(KCP_ROOTDIR); err != nil {
		return nil, err
	}
	go func() {
		command := fmt.Sprintf("%s start --root-directory=%s", KCP_BIN, KCP_ROOTDIR)
		cmd := exec.Command("/bin/sh", "-c", command)
		err := cmd.Start()
		Expect(err).NotTo(HaveOccurred())
		t.kcpPid = cmd.Process.Pid
	}()
	var cfg *rest.Config
	Eventually(func() bool {
		var err error
		cfg, err = getKubeconfig()
		return err == nil
	}, timeout, interval).Should(BeTrue())
	return cfg, nil
}

func (t *KcpTestSetting) TerminateKcp() error {
	if t.kcpPid == 0 {
		return nil
	}
	proc, err := os.FindProcess(t.kcpPid)
	if err == nil {
		err := proc.Kill()
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to terminate KCP (pid=%d)", t.kcpPid))
		}
	}
	return nil
}

func getKubeconfig() (*rest.Config, error) {
	if KCP_ISRUNNING == "true" {
		return getKubeconfigFromFile(KCP_KUBECONFIG)
	}
	return getKubeconfigFromFile(KCP_ROOTDIR + "/admin.kubeconfig")
}
