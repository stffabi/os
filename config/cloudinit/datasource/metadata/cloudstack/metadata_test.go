// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudstack

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/config/cloudinit/datasource/metadata"
	"github.com/rancher/os/config/cloudinit/datasource/metadata/test"
	"github.com/rancher/os/config/cloudinit/pkg"
)

func TestType(t *testing.T) {
	want := "cloudstack-metadata-service"
	if kind := (MetadataService{}).Type(); kind != want {
		t.Fatalf("bad type: want %q, got %q", want, kind)
	}
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root         string
		metadataPath string
		resources    map[string]string
		expect       datasource.Metadata
		clientErr    error
		expectErr    error
	}{
		{
			root:         "/",
			metadataPath: "latest/meta-data/",
			resources: map[string]string{
				"/latest/meta-data/local-hostname": "host",
				"/latest/meta-data/local-ipv4":     "1.2.3.4",
				"/latest/meta-data/public-ipv4":    "5.6.7.8",
				"/latest/meta-data/public-keys":    "key\n",
			},
			expect: datasource.Metadata{
				Hostname:      "host",
				PrivateIPv4:   net.ParseIP("1.2.3.4"),
				PublicIPv4:    net.ParseIP("5.6.7.8"),
				SSHPublicKeys: map[string]string{"0": "key"},
				RootDisk:      nil,
				NetworkConfig: nil,
			},
		},
		{
			root:         "/",
			metadataPath: "latest/meta-data/",
			resources: map[string]string{
				"/latest/meta-data/local-hostname": "host domain another_domain",
				"/latest/meta-data/local-ipv4":     "21.2.3.4",
				"/latest/meta-data/public-ipv4":    "25.6.7.8",
				"/latest/meta-data/public-keys":    "key\n",
			},
			expect: datasource.Metadata{
				Hostname:      "host",
				PrivateIPv4:   net.ParseIP("21.2.3.4"),
				PublicIPv4:    net.ParseIP("25.6.7.8"),
				SSHPublicKeys: map[string]string{"0": "key"},
				RootDisk:      nil,
				NetworkConfig: nil,
			},
		},
		{
			clientErr: pkg.ErrTimeout{Err: fmt.Errorf("test error")},
			expectErr: pkg.ErrTimeout{Err: fmt.Errorf("test error")},
		},
	} {
		service := &MetadataService{metadata.Service{
			Root:         tt.root,
			Client:       &test.HTTPClient{Resources: tt.resources, Err: tt.clientErr},
			MetadataPath: tt.metadataPath,
		}}
		metadata, err := service.FetchMetadata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): \nwant %q, \ngot %q\n", tt.resources, tt.expectErr, err)
		}
		if !reflect.DeepEqual(tt.expect, metadata) {
			t.Fatalf("bad fetch (%q): \nwant %#v, \ngot %#v\n", tt.resources, tt.expect, metadata)
		}
	}
}

func Error(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
