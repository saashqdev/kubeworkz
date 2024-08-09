/*
Copyright 2023 KubeWorkz Authors

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

package options

// Options here for avoid import cycle, remove it as soon as we found better way.
type Options struct {
	KubernetesConfig  string
	AllowPrivileged   bool
	LeaderElect       bool
	WebhookCert       string
	WebhookServerPort int
	// ScoutWaitTimeoutSeconds that heartbeat not receive timeout
	ScoutWaitTimeoutSeconds int
	// ScoutInitialDelaySeconds the time that wait for warden start
	ScoutInitialDelaySeconds int
}