package authutil

type ProfileData struct {
	Name               string `json:"name"`
	Email              string `json:"email"`
	OrganizationDomain string `json:"hd"`
}
