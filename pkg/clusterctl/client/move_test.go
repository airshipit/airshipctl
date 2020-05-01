/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package client

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	bmoapis "github.com/metal3-io/baremetal-operator/pkg/apis"
	bmh "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var bmh1 = &bmh.BareMetalHost{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "metal3.io/v1alpha1",
		Kind:       "BareMetalHost",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "bmh1",
		Namespace: "ns1",
	},
	Spec: bmh.BareMetalHostSpec{
		Online:         false,
		BootMACAddress: "00:2e:30:d7:11:19",
	},
	Status: bmh.BareMetalHostStatus{
		HardwareProfile: "bmh1-hw-profile",
	},
}

var bmh2 = &bmh.BareMetalHost{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "metal3.io/v1alpha1",
		Kind:       "BareMetalHost",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "bmh2",
		Namespace: "ns1",
	},
	Spec: bmh.BareMetalHostSpec{
		Online:         false,
		BootMACAddress: "01:23:45:67:89:ab",
	},
	Status: bmh.BareMetalHostStatus{
		HardwareProfile: "bmh2-hw-profile",
	},
}

func newClientWithBMHObject() client.Client {
	scheme := runtime.NewScheme()
	//nolint:errcheck
	bmoapis.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme, bmh1)
}

func newClientWithTwoBMHObjects() client.Client {
	scheme := runtime.NewScheme()
	//nolint:errcheck
	bmoapis.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme, bmh1, bmh2)
}

func newClientWithNoBMHObject() client.Client {
	scheme := runtime.NewScheme()
	//nolint:errcheck
	bmoapis.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme)
}

func Test_move_getBMHs(t *testing.T) {
	type args struct {
		c         client.Client
		namespace string
	}
	type want struct {
		bmhList bmh.BareMetalHostList
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    want
	}{
		{
			name: "returns a BareMetalHost object",
			args: args{
				c:         newClientWithBMHObject(),
				namespace: "ns1",
			},
			wantErr: false,
			want: want{
				bmhList: bmh.BareMetalHostList{
					Items: []bmh.BareMetalHost{*bmh1},
				},
			},
		},
		{
			name: "returns multiple BareMetalHost object",
			args: args{
				c:         newClientWithTwoBMHObjects(),
				namespace: "ns1",
			},
			wantErr: false,
			want: want{
				bmhList: bmh.BareMetalHostList{
					Items: []bmh.BareMetalHost{*bmh1, *bmh2},
				},
			},
		},
		{
			name: "returns an empty list of BareMetalHost objects",
			args: args{
				c:         newClientWithNoBMHObject(),
				namespace: "ns2",
			},
			wantErr: false,
			want: want{
				bmhList: bmh.BareMetalHostList{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			bmhList, err := getBMHs(context.TODO(), tt.args.c, tt.args.namespace)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(bmhList.Items)).To(BeEquivalentTo(len(tt.want.bmhList.Items)))
		})
	}
}

func Test_move_pauseUnpauseBMHs(t *testing.T) {
	type args struct {
		c         client.Client
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "pause and unpause a single BareMetalHost object",
			args: args{
				c:         newClientWithBMHObject(),
				namespace: "ns1",
			},
			wantErr: false,
		},
		{
			name: "pause and unpause multiple BareMetalHost objects",
			args: args{
				c:         newClientWithTwoBMHObjects(),
				namespace: "ns1",
			},
			wantErr: false,
		},
		{
			name: "pause and unpause should do nothing when there is no BareMetalHost object present",
			args: args{
				c:         newClientWithNoBMHObject(),
				namespace: "ns2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := pauseUnpauseBMHs(context.TODO(), tt.args.c, tt.args.namespace, true)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			bmhList, err := getBMHs(context.TODO(), tt.args.c, tt.args.namespace)
			g.Expect(err).NotTo(HaveOccurred())
			for _, host := range bmhList.Items {
				g.Expect(host.Annotations[bmh.PausedAnnotation]).To(Equal("true"))
			}
			err = pauseUnpauseBMHs(context.TODO(), tt.args.c, tt.args.namespace, false)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			bmhList, err = getBMHs(context.TODO(), tt.args.c, tt.args.namespace)
			g.Expect(err).NotTo(HaveOccurred())
			for _, host := range bmhList.Items {
				_, present := host.Annotations[bmh.PausedAnnotation]
				g.Expect(present).To(Equal(false))
			}
		})
	}
}

var bmh1NoStatus = &bmh.BareMetalHost{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "metal3.io/v1alpha1",
		Kind:       "BareMetalHost",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "bmh1",
		Namespace: "ns1",
	},
	Spec: bmh.BareMetalHostSpec{
		Online:         false,
		BootMACAddress: "00:2e:30:d7:11:19",
	},
}

var bmh2NoStatus = &bmh.BareMetalHost{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "metal3.io/v1alpha1",
		Kind:       "BareMetalHost",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "bmh2",
		Namespace: "ns1",
	},
	Spec: bmh.BareMetalHostSpec{
		Online:         false,
		BootMACAddress: "01:23:45:67:89:ab",
	},
}

func newClientFromCluster() client.Client {
	scheme := runtime.NewScheme()
	//nolint:errcheck
	bmoapis.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme, bmh1, bmh2)
}

func newClientToCluster() client.Client {
	scheme := runtime.NewScheme()
	//nolint:errcheck
	bmoapis.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme, bmh1NoStatus, bmh2NoStatus)
}

func Test_move_copyBMHStatus(t *testing.T) {
	type args struct {
		cFrom     client.Client
		cTo       client.Client
		namespace string
	}
	type want struct {
		bmhList bmh.BareMetalHostList
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    want
	}{
		{
			name: "copies the status field of multiple BareMetalHost objects",
			args: args{
				cFrom:     newClientFromCluster(),
				cTo:       newClientToCluster(),
				namespace: "ns1",
			},
			wantErr: false,
			want: want{
				bmhList: bmh.BareMetalHostList{
					Items: []bmh.BareMetalHost{*bmh1, *bmh2},
				},
			},
		},
		{
			name: "no copy occurs b/c no BareMetalHost objects are present",
			args: args{
				cFrom:     newClientWithNoBMHObject(),
				cTo:       newClientWithNoBMHObject(),
				namespace: "ns1",
			},
			wantErr: false,
			want: want{
				bmhList: bmh.BareMetalHostList{
					Items: []bmh.BareMetalHost{},
				},
			},
		},
		{
			name: "error should occur b/c BareMetalHost does not exist in the source cluster",
			args: args{
				cFrom:     newClientWithNoBMHObject(),
				cTo:       newClientToCluster(),
				namespace: "ns1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := copyBMHStatus(context.TODO(), tt.args.cFrom, tt.args.cTo, tt.args.namespace)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			bmhList, err := getBMHs(context.TODO(), tt.args.cTo, tt.args.namespace)
			g.Expect(err).NotTo(HaveOccurred())
		NEXTHOST:
			for _, host := range bmhList.Items {
				for _, wantHost := range tt.want.bmhList.Items {
					if host.Name == wantHost.Name {
						g.Expect(host.Status.HardwareProfile).To(Equal(wantHost.Status.HardwareProfile))
						continue NEXTHOST
					}
				}
				t.Errorf("unexpected host %s", host.Name)
			}
		})
	}
}
