// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestInvitation(t *testing.T) {
	InitConfig(TestCasdoorEndpoint, TestClientId, TestClientSecret, TestJwtPublicKey, TestCasdoorOrganization, TestCasdoorApplication)

	name := getRandomName("unit_test_invitation")
	code := "TEST1234"
	// Test invitation object
	invitation := &Invitation{
		Owner:       TestCasdoorOrganization,
		Name:        name,
		CreatedTime: time.Now().Format(time.RFC3339),
		DisplayName: "Test Invitation",
		Code:        code,
		DefaultCode: code,
		Quota:       10,
		UsedCount:   0,
		Application: TestCasdoorApplication,
		Email:       "test@example.com",
		SignupGroup: "test-group",
		State:       "Active",
	}

	// Test AddInvitation
	_, err := AddInvitation(invitation)
	if err != nil {
		t.Fatalf("Failed to add invitation: %v", err)
	}

	// Test GetInvitation
	invitation2, err := GetInvitation(name)
	if err != nil {
		t.Fatalf("Failed to get invitation: %v", err)
	}
	if invitation2.Code != invitation.Code {
		t.Fatalf("Retrieved invitation does not match added invitation")
	}

	// Test GetInvitations
	invitations, err := GetInvitations()
	if err != nil {
		t.Fatalf("Failed to get invitations: %v", err)
	}
	if len(invitations) == 0 {
		t.Fatalf("No invitations found")
	}

	// Test UpdateInvitation
	invitation2.State = "Suspended"
	_, err = UpdateInvitation(invitation2)
	if err != nil {
		t.Fatalf("Failed to update invitation: %v", err)
	}

	// Test UpdateInvitation to Active to check getInvitationInfo
	invitation2.State = "Active"
	_, err = UpdateInvitation(invitation2)
	if err != nil {
		t.Fatalf("Failed to update invitation: %v", err)
	}

	// Test GetInvitationInfo
	invitation, err = GetInvitationInfo(code, "app-casbin")
	if err != nil {
		t.Fatalf("Failed to get invitation info by code: %v", err)
	}
	if invitation == nil {
		t.Fatalf("Invitation not found by code")
	}

	// Test DeleteInvitation
	_, err = DeleteInvitation(invitation2)
	if err != nil {
		t.Fatalf("Failed to delete invitation: %v", err)
	}

	// Verify deletion
	deletedInvitation, err := GetInvitation(name)
	if err == nil && deletedInvitation != nil {
		t.Fatalf("Failed to delete invitation, it still exists")
	}
}

func TestSendInvitationRequest(t *testing.T) {
	var gotPath, gotId string
	var gotDestinations []string
	response := `{"status":"ok","msg":""}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotId = r.URL.Query().Get("id")

		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotDestinations); err != nil {
			t.Errorf("Failed to parse the body %q: %v", body, err)
		}

		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	c := NewClient(server.URL, "clientId", "clientSecret", "", "casbin", "app")

	destinations := []string{"a@example.com", "b@example.com"}
	ok, err := c.SendInvitation("invitation1", destinations)
	if err != nil {
		t.Fatalf("Failed to send invitation: %v", err)
	}
	if !ok {
		t.Fatalf("Expected the invitation to be sent")
	}
	if gotPath != "/api/send-invitation" || gotId != "casbin/invitation1" || !reflect.DeepEqual(gotDestinations, destinations) {
		t.Fatalf("Unexpected request: path=%s id=%s destinations=%v", gotPath, gotId, gotDestinations)
	}

	response = `{"status":"error","msg":"mail failed"}`
	ok, err = c.SendInvitation("invitation1", destinations)
	if err == nil || err.Error() != "mail failed" || ok {
		t.Fatalf("Expected the error from Casdoor, got %v and %v", ok, err)
	}
}

func TestSendInvitation(t *testing.T) {
	InitConfig(TestCasdoorEndpoint, TestClientId, TestClientSecret, TestJwtPublicKey, TestCasdoorOrganization, TestCasdoorApplication)

	// the invitation doesn't exist
	_, err := SendInvitation(getRandomName("not_exist"), []string{"test@example.com"})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("Expected the invitation to not exist, got %v", err)
	}

	// the invitation exists, so Casdoor goes on to send the email, which can't be delivered by
	// the unreachable SMTP server of the test application
	name := getRandomName("unit_test_invitation")
	invitation := &Invitation{
		Owner:       TestCasdoorOrganization,
		Name:        name,
		CreatedTime: time.Now().Format(time.RFC3339),
		Code:        "SENDCODE1234",
		DefaultCode: "SENDCODE1234",
		Quota:       1,
		Application: TestCasdoorApplication,
		State:       "Active",
	}
	if _, err = AddInvitation(invitation); err != nil {
		t.Fatalf("Failed to add invitation: %v", err)
	}
	defer func() {
		_, _ = DeleteInvitation(invitation)
	}()

	ok, err := SendInvitation(name, []string{"test@example.com"})
	if err == nil || ok {
		t.Fatalf("Expected the email to fail, got %v and %v", ok, err)
	}
	if strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("Expected the invitation to be found, got %v", err)
	}
}
