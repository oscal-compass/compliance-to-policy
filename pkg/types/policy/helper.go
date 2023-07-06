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

package policy

type policyInterface interface {
	getNamespace() string
	getName() string
	getAnnotation() map[string]string
}

func (p *Policy) getNamespace() string {
	return p.Namespace
}

func (p *Policy) getName() string {
	return p.Name
}

func (p *Policy) getAnnotation() map[string]string {
	return p.Annotations
}

func (p *PolicySet) getNamespace() string {
	return p.Namespace
}

func (p *PolicySet) getName() string {
	return p.Name
}

func (p *PolicySet) getAnnotation() map[string]string {
	return p.Annotations
}

func FindByNamespaceName[T policyInterface](list []T, namespace string, name string) T {
	var result T
	for _, item := range list {
		if item.getNamespace() == namespace && item.getName() == name {
			return item
		}
	}
	return result
}

func FindByNamespaceAnnotation[T policyInterface](list []T, namespace string, annotationName string, annotationValue string) T {
	var result T
	for _, item := range list {
		if item.getNamespace() == namespace {
			an, ok := item.getAnnotation()[annotationName]
			if ok {
				if an == annotationValue {
					return item
				}
			}
		}
	}
	return result
}
