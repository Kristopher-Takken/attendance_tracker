// models.go
package main

type User struct {
	Username   string `json:"username"`
	Password   string `json:"password"` // Store as a hash
	Admin      bool   `json:"admin"`
	SuperAdmin bool   `json:"super_admin"`
}

type Attendance struct {
	Datetime     string `json:"datetime"` // YYYY-MM-DD: hh:mm:ss
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Reason       string `json:"reason"`
	LicenceState string `json:"licence_state"`
	User         string `json:"user"`
}
