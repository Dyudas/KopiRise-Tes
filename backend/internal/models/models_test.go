package models

import "testing"

func TestUserRoles(t *testing.T) {
	u := &User{
		ID:    "usr-01",
		Roles: []string{RoleAdmin, RoleManager},
	}

	if !u.HasRole(RoleAdmin) {
		t.Error("User seharusnya memiliki peran admin")
	}
	if !u.HasRole(RoleManager) {
		t.Error("User seharusnya memiliki peran manager")
	}
	if u.HasRole(RoleCashier) {
		t.Error("User seharusnya TIDAK memiliki peran cashier")
	}

	if !u.HasAnyRole(RoleBarista, RoleAdmin) {
		t.Error("HasAnyRole seharusnya bernilai true jika setidaknya satu peran cocok")
	}
	if u.HasAnyRole(RoleWaiter, RoleCustomer) {
		t.Error("HasAnyRole seharusnya bernilai false jika tidak ada peran yang cocok")
	}
}
