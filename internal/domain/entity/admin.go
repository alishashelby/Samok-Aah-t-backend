package entity

type Admin struct {
	ID          int64
	AuthID      int64
	Permissions map[string]bool
}

func NewAdmin(authID int64, permissions map[string]bool) *Admin {
	return &Admin{
		AuthID:      authID,
		Permissions: permissions,
	}
}
