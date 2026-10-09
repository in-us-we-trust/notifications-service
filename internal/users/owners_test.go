package users

import "testing"

func TestOwners(t *testing.T) {
	contacts := []User{
		{UserID: "owner", IsOwner: true},
		{UserID: "guest", IsOwner: false},
		{UserID: "other-owner", IsOwner: true},
	}

	owners := Owners(contacts)
	if len(owners) != 2 {
		t.Fatalf("len = %d", len(owners))
	}
	if owners[0].UserID != "owner" || owners[1].UserID != "other-owner" {
		t.Fatalf("owners = %+v", owners)
	}
}

func TestOwnersEmptyWhenNobodyOwnsTheDevice(t *testing.T) {
	owners := Owners([]User{{UserID: "guest", IsOwner: false}})
	if len(owners) != 0 {
		t.Fatalf("len = %d", len(owners))
	}
}
