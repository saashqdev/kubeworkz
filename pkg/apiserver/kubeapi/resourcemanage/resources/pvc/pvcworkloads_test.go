/*
Copyright 2024 KubeWorkz Authors

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

package pvc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/pod"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/multicluster"
	"github.com/saashqdev/kubeworkz/pkg/multicluster/client/fake"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
)

var _ = Describe("Pvc", func() {
	var (
		ns      = "namespace-test"
		pvcName = "pvc-name"
		pod1    corev1.Pod
		pod2    corev1.Pod
		podList corev1.PodList
	)
	BeforeEach(func() {
		pvc := corev1.PersistentVolumeClaimVolumeSource{ClaimName: pvcName}
		v := corev1.Volume{VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &pvc}}
		pod1 = corev1.Pod{
			TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: ns},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{},
			},
		}
		pod2 = corev1.Pod{
			TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: ns},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{v},
			},
		}
		podList = corev1.PodList{
			Items: []corev1.Pod{pod1, pod2},
		}

	})
	JustBeforeEach(func() {
		scheme := runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		opts := &fake.Options{
			Scheme:               scheme,
			Objs:                 []client.Object{},
			ClientSetRuntimeObjs: []runtime.Object{},
			Lists:                []client.ObjectList{&podList},
		}
		multicluster.InitFakeMultiClusterMgrWithOpts(opts)
		clients.InitCubeClientSetWithOpts(nil)
	})

	It("test get pvc workloads (pod which used this pvc)", func() {
		client := clients.Interface().Kubernetes(constants.LocalCluster)
		Expect(client).NotTo(BeNil())
		pvc := NewPvc(client, ns, nil)
		ret, err := pvc.getPvcWorkloads(pvcName)
		Expect(err).To(BeNil())
		Expect(ret.Object["total"]).To(Equal(1))
		pods := ret.Object["pods"].([]pod.ExtendPod)
		s := pods[0].Name
		Expect(s).To(Equal("pod2"))
	})
})