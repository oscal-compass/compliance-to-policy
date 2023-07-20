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

type KubernetesObject interface {
	GetNamespace() string
	GetName() string
	GetAnnotation() map[string]string
	GetLabel() map[string]string
}

func FilterByAnnotation[T KubernetesObject](list []T, annotationName string, annotationValue string) []T {
	filtered := []T{}
	for _, item := range list {
		an, ok := item.GetAnnotation()[annotationName]
		if ok {
			if an == annotationValue {
				filtered = append(filtered, item)
			}
		}
	}
	return filtered
}

func FindByNamespaceName[T KubernetesObject](list []T, namespace string, name string) T {
	var result T
	for _, item := range list {
		if item.GetNamespace() == namespace && item.GetName() == name {
			return item
		}
	}
	return result
}

func FindByNamespaceAnnotation[T KubernetesObject](list []T, namespace string, annotationName string, annotationValue string) T {
	var result T
	for _, item := range list {
		if item.GetNamespace() == namespace {
			an, ok := item.GetAnnotation()[annotationName]
			if ok {
				if an == annotationValue {
					return item
				}
			}
		}
	}
	return result
}

func FindByNamespaceLabel[T KubernetesObject](list []T, namespace string, labelName string, labelValue string) T {
	var result T
	for _, item := range list {
		if item.GetNamespace() == namespace {
			an, ok := item.GetLabel()[labelName]
			if ok {
				if an == labelValue {
					return item
				}
			}
		}
	}
	return result
}
