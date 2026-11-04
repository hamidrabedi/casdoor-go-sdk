// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casdoorsdk

import (
	"encoding/json"
	"testing"
)

// Updates send the whole object and the server writes every column, so a field the
// SDK doesn't declare is blanked by any read-modify-write through the SDK.
func assertRoundTrip(t *testing.T, serverJson string, object interface{}) {
	t.Helper()

	if err := json.Unmarshal([]byte(serverJson), object); err != nil {
		t.Fatal(err)
	}
	sent, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}

	var want, got map[string]interface{}
	if err := json.Unmarshal([]byte(serverJson), &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(sent, &got); err != nil {
		t.Fatal(err)
	}
	for key, value := range want {
		if !jsonEqual(got[key], value) {
			t.Errorf("field %s: sent %v, want %v", key, got[key], value)
		}
	}
}

func jsonEqual(a, b interface{}) bool {
	first, _ := json.Marshal(a)
	second, _ := json.Marshal(b)
	return string(first) == string(second)
}

func TestSyncerKeepsServerFields(t *testing.T) {
	assertRoundTrip(t, `{
		"sslMode": "verify-full",
		"sslRootCert": "root",
		"sslCert": "cert",
		"sslKey": "***",
		"sshHostKey": "host-key",
		"lastSyncTime": "2026-11-04T10:00:00Z"
	}`, &Syncer{})
}

func TestOrganizationKeepsServerFields(t *testing.T) {
	assertRoundTrip(t, `{
		"trustedDbHosts": "postgres:9010",
		"useOrganizationIssuer": true,
		"idJagIssuers": [{"issuer": "i", "jwksUri": "j", "subject": "s", "audience": "a", "user": "u"}],
		"enableLdapPassword": true,
		"enableRadius": true,
		"disableRadiusMfa": true
	}`, &Organization{})
}
