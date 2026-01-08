package auth

// Role represents a user role in the system
type Role string

const (
	RoleOwnerAdmin Role = "owner_admin"
	RoleManager    Role = "manager"
	RoleAccountant Role = "accountant"
	RoleViewer     Role = "viewer"
)

// AllRoles returns all valid roles
func AllRoles() []Role {
	return []Role{RoleOwnerAdmin, RoleManager, RoleAccountant, RoleViewer}
}

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	for _, valid := range AllRoles() {
		if r == valid {
			return true
		}
	}
	return false
}

// CanImport returns true if the role can perform imports
func (r Role) CanImport() bool {
	return r == RoleOwnerAdmin || r == RoleAccountant
}

// CanExport returns true if the role can perform exports
func (r Role) CanExport() bool {
	return r == RoleOwnerAdmin || r == RoleManager || r == RoleAccountant
}

// CanViewDashboard returns true if the role can view the dashboard
func (r Role) CanViewDashboard() bool {
	return true // All roles can view dashboard
}

// CanManageUsers returns true if the role can manage users
func (r Role) CanManageUsers() bool {
	return r == RoleOwnerAdmin
}

// CanEditMappings returns true if the role can edit mapping profiles
func (r Role) CanEditMappings() bool {
	return r == RoleOwnerAdmin || r == RoleAccountant
}
