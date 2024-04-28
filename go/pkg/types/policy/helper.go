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

func (p *Policy) GetNamespace() string {
	return p.Namespace
}

func (p *Policy) GetName() string {
	return p.Name
}

func (p *Policy) GetAnnotation() map[string]string {
	return p.Annotations
}

func (p *Policy) GetLabel() map[string]string {
	return p.Labels
}

func (p *PolicySet) GetNamespace() string {
	return p.Namespace
}

func (p *PolicySet) GetName() string {
	return p.Name
}

func (p *PolicySet) GetAnnotation() map[string]string {
	return p.Annotations
}

func (p *PolicySet) GetLabel() map[string]string {
	return p.Labels
}
