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

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	c2pv1alpha1 "github.com/oscal-compass/compliance-to-policy/go/api/v1alpha1"
	edge "github.com/oscal-compass/compliance-to-policy/go/controllers/edge.kcp.io/v1alpha1"
	"github.com/oscal-compass/compliance-to-policy/go/controllers/utils/kcpclient"
	"github.com/oscal-compass/compliance-to-policy/go/pkg"
	"github.com/oscal-compass/compliance-to-policy/go/pkg/oscal"
	internalcompliance "github.com/oscal-compass/compliance-to-policy/go/pkg/types/internalcompliance"
	typesoscal "github.com/oscal-compass/compliance-to-policy/go/pkg/types/oscal"
	cd "github.com/oscal-compass/compliance-to-policy/go/pkg/types/oscal/componentdefinition"
	"gopkg.in/src-d/go-git.v4"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logger logr.Logger = ctrl.Log.WithName("controller-common-utils")

var gitRepoCache map[string]string = map[string]string{}

func HandleError(logger logr.Logger, err error, message string) (ctrl.Result, error) {
	logger.Error(err, message)
	return ctrl.Result{Requeue: false}, nil
}

func ConvertIntComplianceToCompliance(intCompliance internalcompliance.Compliance) c2pv1alpha1.Compliance {
	intStandard := intCompliance.Standard
	intCategories := intStandard.Categories
	categories := []c2pv1alpha1.Category{}
	for _, intCategory := range intCategories {
		controls := []c2pv1alpha1.Control{}
		for _, intControl := range intCategory.Controls {
			controls = append(controls, c2pv1alpha1.Control{
				Name:        intControl.Name,
				ControlRefs: intControl.ControlRefs,
			})
		}
		categories = append(categories, c2pv1alpha1.Category{
			Name:     intCategory.Name,
			Controls: controls,
		})
	}
	standard := c2pv1alpha1.Standard{
		Name:       intStandard.Name,
		Categories: categories,
	}
	return c2pv1alpha1.Compliance{
		Standard: standard,
	}
}

func ConvertComplianceToIntCompliance(compliance c2pv1alpha1.Compliance) internalcompliance.Compliance {
	standard := internalcompliance.Standard{
		Name: compliance.Standard.Name,
	}
	categories := []internalcompliance.Category{}
	for _, category := range compliance.Standard.Categories {
		intCategory := internalcompliance.Category{
			Name: category.Name,
		}
		intControls := []internalcompliance.Control{}
		for _, control := range category.Controls {
			inControl := internalcompliance.Control{
				Name:        control.Name,
				ControlRefs: control.ControlRefs,
			}
			intControls = append(intControls, inControl)
		}
		intCategory.Controls = intControls
		categories = append(categories, intCategory)
	}
	standard.Categories = categories
	return internalcompliance.Compliance{
		Standard: standard,
	}
}

type crComposit struct {
	ControlReference    c2pv1alpha1.ControlReference
	ControlReferenceKcp c2pv1alpha1.ControlReferenceKcp
	Catalog             *typesoscal.CatalogRoot
	Profile             *typesoscal.ProfileRoot
	ComponentDefinition *cd.ComponentDefinitionRoot
}

func MakeControlReference(
	tempDir string,
	compDeploy c2pv1alpha1.ComplianceDeployment,
) (crComposit, error) {

	var cr c2pv1alpha1.ControlReference
	var _crComposit crComposit

	intCompliance, summary, _crComposit, err := makeControlReference(tempDir, compDeploy)
	if err != nil {
		return _crComposit, err
	}

	compliance := ConvertIntComplianceToCompliance(intCompliance)
	target := c2pv1alpha1.ControlReferenceTarget{
		Namespace: compDeploy.Spec.Target.Namespace,
	}
	cr = c2pv1alpha1.ControlReference{
		ObjectMeta: v1.ObjectMeta{
			Name:      compDeploy.Name,
			Namespace: compDeploy.Namespace,
		},
		Spec: c2pv1alpha1.ControlReferenceSpec{
			Compliance:      compliance,
			Target:          target,
			PolicyResources: compDeploy.Spec.PolicyResources,
			Summary:         summary,
		},
	}

	crkcp := c2pv1alpha1.ControlReferenceKcp{
		ObjectMeta: v1.ObjectMeta{
			Name:      compDeploy.Name,
			Namespace: compDeploy.Namespace,
		},
		Spec: c2pv1alpha1.ControlReferenceKcpSpec{
			Compliance:           compliance,
			ComplianceDeployment: compDeploy.Spec,
			Summary:              summary,
		},
	}

	_crComposit.ControlReference = cr
	_crComposit.ControlReferenceKcp = crkcp
	return _crComposit, nil
}

func makeControlReference(tempDir string, compDeploy c2pv1alpha1.ComplianceDeployment) (internalcompliance.Compliance, map[string]string, crComposit, error) {
	var intCompliance internalcompliance.Compliance
	var summary map[string]string
	var _crComposit crComposit

	logger.Info(fmt.Sprintf("Component-definition is loaded from %s", compDeploy.Spec.Compliance.ComponentDefinition.Url))
	var cdobj cd.ComponentDefinitionRoot
	if err := loadFromGit(compDeploy.Spec.Compliance.ComponentDefinition.Url, tempDir, &cdobj); err != nil {
		logger.Error(err, "Failed to load component-definition")
		return intCompliance, summary, _crComposit, err
	}

	logger.Info(fmt.Sprintf("Catalog is loaded from %s", compDeploy.Spec.Compliance.Catalog.Url))
	var catalogObj typesoscal.CatalogRoot
	if err := loadFromWeb(compDeploy.Spec.Compliance.Catalog.Url, &catalogObj); err != nil {
		logger.Error(err, "Failed to load catalog")
		return intCompliance, summary, _crComposit, err
	}

	logger.Info(fmt.Sprintf("Profile is loaded from %s", compDeploy.Spec.Compliance.Profile.Url))
	var profileObj typesoscal.ProfileRoot
	if err := loadFromWeb(compDeploy.Spec.Compliance.Profile.Url, &profileObj); err != nil {
		logger.Error(err, "Failed to load profile")
		return intCompliance, summary, _crComposit, err
	}

	profiledCd := oscal.IntersectProfileWithCD(cdobj.ComponentDefinition, profileObj.Profile)
	intCompliance = oscal.MakeInternalCompliance(catalogObj.Catalog, profileObj.Profile, profiledCd)

	summary = logControlIds(logger, profileObj.Profile, cdobj.ComponentDefinition, intCompliance)
	summary["name"] = profileObj.Metadata.Title
	summary["compliance-definition-name"] = compDeploy.Name
	summary["compliance-definition-namespace"] = compDeploy.Spec.Target.Namespace
	summary["catalog"] = compDeploy.Spec.Compliance.Catalog.Url
	summary["profile"] = compDeploy.Spec.Compliance.Profile.Url
	summary["component-definition"] = compDeploy.Spec.Compliance.ComponentDefinition.Url

	logger.Info("Required policies are")
	for _, category := range intCompliance.Standard.Categories {
		for _, control := range category.Controls {
			logger.Info(fmt.Sprintf("- %s: %v", control.Name, control.ControlRefs))
		}
	}

	_crComposit = crComposit{
		Catalog:             &catalogObj,
		Profile:             &profileObj,
		ComponentDefinition: &cdobj,
	}

	return intCompliance, summary, _crComposit, nil
}

func loadFromWeb(url string, out interface{}) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	if u.Scheme == "local" {
		if err := pkg.LoadJsonFileToObject(u.Path, out); err != nil {
			return fmt.Errorf("Failed to marshal %s in local directory", u.Path)
		}
		return nil
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Failed to initialize http client for %s", url)
	}

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to get %s", url)
	}
	defer resp.Body.Close()

	byteArray, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to serialize body %s", url)
	}

	err = json.Unmarshal(byteArray, out)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal %s", url)
	}
	return nil
}

func loadFromGit(url string, tempDir string, out interface{}) error {
	u, err := neturl.Parse(url)
	if err != nil {
		return err
	}
	paths := strings.Split(u.Path, "/")
	repoUrl := fmt.Sprintf("%s://%s/%s/%s", u.Scheme, u.Host, paths[1], paths[2])
	path := strings.Join(paths[3:], "/")

	if u.Scheme == "local" {
		if err := pkg.LoadJsonFileToObject(u.Path, out); err != nil {
			return fmt.Errorf("Failed to marshal %s in local directory", u.Path)
		}
		return nil
	}
	repoDir, _, err := GitClone(repoUrl, tempDir)
	if err != nil {
		return fmt.Errorf("Failed to clone %s", repoUrl)
	}
	if err := pkg.LoadJsonFileToObject(repoDir+"/"+path, out); err != nil {
		return fmt.Errorf("Failed to marshal %s", repoDir+path)
	}
	return nil
}

func GitClone(url string, tmpdir string) (string, string, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return "", "", err
	}
	paths := strings.Split(u.Path, "/")
	repoUrl := fmt.Sprintf("%s://%s/%s/%s", u.Scheme, u.Host, paths[1], paths[2])
	path := strings.Join(paths[3:], "/")
	if u.Scheme == "local" {
		return u.Path, "", nil
	}
	rootDir, err := gitClone(repoUrl, tmpdir)
	return rootDir, path, err
}

func gitClone(url string, tmpdir string) (string, error) {
	username := os.Getenv("username")
	token := os.Getenv("token")
	dir, ok := gitRepoCache[url]
	if !ok {
		dir, err := os.MkdirTemp(tmpdir, "tmp-")
		if err != nil {
			return "", err
		}
		cloneOption := &git.CloneOptions{
			URL: url,
		}
		if username != "" && token != "" {
			logger.Info("Git Clone with Auth given by 'username' and 'token' in environment variables ")
			cloneOption.Auth = &githttp.BasicAuth{Username: username, Password: token}
		}
		if _, err := git.PlainClone(dir, false, cloneOption); err != nil {
			return "", err
		}
		gitRepoCache[url] = dir
		return dir, nil
	}
	return dir, nil
}

func logControlIds(logger logr.Logger, profile typesoscal.Profile, compDef cd.ComponentDefinition, intCompliance internalcompliance.Compliance) map[string]string {
	controlIdsInProfile := []string{}
	for _, profileImport := range profile.Imports {
		for _, includeControl := range profileImport.IncludeControls {
			controlIdsInProfile = append(controlIdsInProfile, includeControl.WithIds...)
		}
	}
	controlIdsInCD := []string{}
	for _, category := range intCompliance.Standard.Categories {
		for _, control := range category.Controls {
			controlIdsInCD = append(controlIdsInCD, control.Name)
		}
	}
	excludedControlIds := []string{}
	findId := func(id string) bool {
		for _, idInCD := range controlIdsInCD {
			if id == idInCD {
				return true
			}
		}
		return false
	}
	for _, idInProfile := range controlIdsInProfile {
		if !findId(idInProfile) {
			excludedControlIds = append(excludedControlIds, idInProfile)
		}
	}
	logger.Info("Profile requires controls")
	logger.Info(fmt.Sprintf("%v", controlIdsInProfile))
	logger.Info("Component Definition describes implemented controls")
	logger.Info(fmt.Sprintf("%v", controlIdsInCD))
	logger.Info("Skipped controls")
	logger.Info(fmt.Sprintf("%v", excludedControlIds))

	summary := map[string]string{}
	summary["controlIdsInProfile"] = strconv.Itoa(len(controlIdsInProfile))
	summary["controlIdsInCD"] = strconv.Itoa(len(controlIdsInCD))
	summary["excludedControlIds"] = strconv.Itoa(len(excludedControlIds))

	return summary
}

type Workspace struct {
	Name           string
	SyncTargetName string
}

func GetWorkspaces(ctx context.Context, cfg rest.Config, workspace string) ([]Workspace, error) {
	kcpClient, err := kcpclient.NewKcpClient(cfg, workspace)
	if err != nil {
		return nil, err
	}
	wsDyClient, err := kcpClient.GetDyClient("tenancy.kcp.io", "Workspace", "v1alpha1")
	if err != nil {
		return nil, err
	}
	workspaces := []Workspace{}
	unstList, err := wsDyClient.List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, unstObj := range unstList.Items {
		syncTargetname, ok := unstObj.GetAnnotations()["edge.kcp.io/sync-target-name"]
		if !ok {
			syncTargetname = ""
		}
		ws := Workspace{
			Name:           workspace + ":" + unstObj.GetName(),
			SyncTargetName: syncTargetname,
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func GetSts(
	ctx context.Context,
	cfg rest.Config,
	wmwName string,
	epname string,
) (edge.SinglePlacementSlice, error) {

	var sps edge.SinglePlacementSlice
	kcpClient, err := kcpclient.NewKcpClient(cfg, wmwName)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to get kcpClient %s", epname))
		return sps, err
	}
	spsClient, err := kcpClient.GetDyClient("edge.kcp.io", "SinglePlacementSlice", "v1alpha1")
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to get spsClient %s", epname))
		return sps, err
	}
	spsUnst, err := spsClient.Get(ctx, epname, v1.GetOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to get sps %s", epname))
		return sps, err
	}
	if err := pkg.ToK8sTypedObject(spsUnst, &sps); err != nil {
		logger.Error(err, fmt.Sprintf("Failed to convert unstObj to sps %s", epname))
		return sps, err
	}
	return sps, nil
}

func CreateOrUpdate[T client.Object](ctx context.Context, r client.Client, obj T, fetched T) error {
	logger := log.FromContext(ctx)
	nsName := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	err := r.Get(ctx, nsName, fetched, &client.GetOptions{})
	if errors.IsNotFound(err) {
		if err := r.Create(ctx, obj, &client.CreateOptions{}); err != nil {
			logger.Error(err, "Failed to create")
			return err
		}
	} else if err == nil {
		obj.SetUID(fetched.GetUID())
		obj.SetResourceVersion(fetched.GetResourceVersion())
		if err := r.Update(ctx, obj, &client.UpdateOptions{}); err != nil {
			logger.Error(err, "Failed to update")
			return err
		}
	} else {
		logger.Error(err, "Failed to get")
		return err
	}
	return nil
}
