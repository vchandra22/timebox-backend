package service

import "testing"

func TestCloudinarySignatureSortsParams(t *testing.T) {
	svc := newCollaborationService(nil, nil, CollaborationOptions{APISecret: "secret"})

	got := svc.cloudinarySignature(map[string]string{
		"timestamp": "1783049700",
		"folder":    "timebox-space/development/ws/task",
		"public_id": "task/task-1",
	})
	want := "c32f257377bc324b42a74dba60864e52b319dd33"
	if got != want {
		t.Fatalf("signature = %s, want %s", got, want)
	}
}
