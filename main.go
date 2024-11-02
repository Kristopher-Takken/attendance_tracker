// main.go
package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {

	// Initialize the default superuser if no users exist
	initDefaultSuperUser()

	router := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// Load HTML templates and serve static files
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	// Define routes for authentication
	router.GET("/login", showLoginForm)
	router.POST("/login", loginUser)
	router.GET("/logout", logoutUser)

	router.Use(requireLoginMiddleware)
	router.GET("/", showHomePage)

	// Define routes for user management
	userRoutes := router.Group("/users")
	userRoutes.Use(requireAdmin)
	{
		userRoutes.GET("/", listUsers) // Change from `router.GET` to `userRoutes.GET`
		userRoutes.GET("/new", newUserForm)
		userRoutes.POST("/new", addUser)
		userRoutes.GET("/edit/:username", editUserForm)
		userRoutes.POST("/edit/:username", updateUser)
		userRoutes.POST("/delete/:username", deleteUser)
	}

	// Group routes requiring superAdmin access
	superAdminRoutes := router.Group("/attendance")
	superAdminRoutes.Use(requireSuperAdmin)
	{
		superAdminRoutes.POST("/delete_all", deleteAllAttendance)
	}

	// Group routes requiring admin access
	adminRoutes := router.Group("/attendance")
	adminRoutes.Use(requireAdmin)
	{
		adminRoutes.GET("/download", downloadAttendance)
	}

	// Define routes for attendance management
	router.GET("/attendance", listAttendance)
	router.GET("/attendance/new", newAttendanceForm)
	router.POST("/attendance/new", requireLogin, addAttendance)
	//router.GET("/attendance/download", downloadAttendance)
	//router.POST("/attendance/delete_all", deleteAllAttendance)

	// Start the server
	router.Run(":80")
}

func requireLoginMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("username") == nil {
		// If user is not logged in, redirect to the login page
		c.Redirect(http.StatusFound, "/login")
		c.Abort() // Prevents further processing of the request
		return
	}
	c.Next() // Continue to the next handler if logged in
}

func showHomePage(c *gin.Context) {
	session := sessions.Default(c)
	username, _ := session.Get("username").(string)
	isAdmin, _ := session.Get("admin").(bool)

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"username": username,
		"isAdmin":  isAdmin,
	})
}

// initDefaultSuperUser checks if there are any users in the system.
// If no users are found, it creates a default superuser with
// username "admin" and password "admin".
func initDefaultSuperUser() {
	users, err := GetAllUsers()
	if err != nil {
		fmt.Println("Error reading users:", err)
		return
	}

	// Check if a superuser already exists
	superuserExists := false
	for _, user := range users {
		if user.SuperAdmin {
			superuserExists = true
			break
		}
	}

	// If no superuser exists, create a default superuser
	if !superuserExists {
		defaultUser := User{
			Username:   "admin",
			Password:   hashPassword("admin"),
			Admin:      true,
			SuperAdmin: true,
		}
		err := AddUser(defaultUser)
		if err != nil {
			fmt.Println("Error creating default superuser:", err)
		} else {
			fmt.Println("Default superuser created: username=admin, password=admin")
		}
	}
}

func requireSuperAdmin(c *gin.Context) {
	// Retrieve the username from the session
	session := sessions.Default(c)
	username, ok := session.Get("username").(string)
	if !ok || username == "" {
		// No username in the session, deny access
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{
			"message": "Access denied: insufficient privileges",
		})
		c.Abort()
		return
	}

	// Load all users from the JSON file
	users, err := GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"message": "Error retrieving user data",
		})
		c.Abort()
		return
	}

	// Find the user in the list and check for superAdmin privileges
	var isSuperAdmin bool
	for _, user := range users {
		if user.Username == username {
			isSuperAdmin = user.SuperAdmin
			break
		}
	}

	// Check if the user has superAdmin privileges
	if !isSuperAdmin {
		// User does not have superAdmin privileges
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{
			"message": "Access denied: insufficient privileges",
		})
		c.Abort()
		return
	}

	c.Next() // Continue to the next handler if the user is a superAdmin
}

func requireAdmin(c *gin.Context) {
	session := sessions.Default(c)
	username, ok := session.Get("username").(string)
	if !ok || username == "" {
		// If there's no username in the session, deny access
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{
			"message": "Access denied: insufficient privileges",
		})
		c.Abort()
		return
	}

	// Load all users from the JSON file
	users, err := GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"message": "Error retrieving user data",
		})
		c.Abort()
		return
	}

	// Find the user in the list and check their privileges
	var isAdmin, isSuperAdmin bool
	for _, user := range users {
		if user.Username == username {
			isAdmin = user.Admin
			isSuperAdmin = user.SuperAdmin
			break
		}
	}

	// Check if the user has either admin or superAdmin privileges
	if !isAdmin && !isSuperAdmin {
		// User does not have sufficient privileges
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{
			"message": "Access denied: insufficient privileges",
		})
		c.Abort()
		return
	}

	c.Next() // Continue to the next handler if the user is an admin or superAdmin
}

// showLoginForm displays the login form.
func showLoginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", nil)
}

func requireLogin(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("username") == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "User not authenticated"})
		c.Abort()
		return
	}
	c.Next()
}

func loginUser(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	users, _ := GetAllUsers()
	for _, user := range users {
		if user.Username == username && verifyPassword(user.Password, password) {
			// Set admin and superAdmin status based on user role
			c.Set("admin", user.Admin)
			c.Set("superAdmin", user.SuperAdmin)

			// Store the username and roles in the session
			session := sessions.Default(c)
			session.Set("username", username)
			session.Set("admin", user.Admin)
			session.Set("superAdmin", user.SuperAdmin)
			session.Save()

			c.Redirect(http.StatusFound, "/")
			return
		}
	}
	c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "Invalid credentials"})
}

// logoutUser handles user logout by clearing session or cookies.
func logoutUser(c *gin.Context) {
	// Retrieve the session
	session := sessions.Default(c)

	// Clear the session data
	session.Clear()

	// Save the session to persist changes
	session.Save()

	// Redirect to the login page
	c.Redirect(http.StatusFound, "/login")
}

// listUsers displays all users with options to edit and delete.
func listUsers(c *gin.Context) {
	session := sessions.Default(c)
	username, ok := session.Get("username").(string)

	if !ok || username == "" {
		username = "Guest" // Fallback if username is missing
	}

	users, err := GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"message":  "Failed to load users",
			"username": username,
		})
		return
	}
	c.HTML(http.StatusOK, "users.tmpl", gin.H{
		"users":    users,
		"username": username,
	})
}

// newUserForm displays the form for creating a new user.
func newUserForm(c *gin.Context) {
	session := sessions.Default(c)
	username, ok := session.Get("username").(string)

	if !ok || username == "" {
		username = "Guest" // Fallback if username is missing
	}
	c.HTML(http.StatusOK, "user_form.tmpl", gin.H{
		"action":   "/users/new",
		"username": username,
	})
}

// addUser handles adding a new user to the system.
func addUser(c *gin.Context) {
	username := c.PostForm("username")
	password := hashPassword(c.PostForm("password"))
	admin := c.PostForm("admin") == "on"
	superAdmin := c.PostForm("super_admin") == "on"

	user := User{Username: username, Password: password, Admin: admin, SuperAdmin: superAdmin}
	err := AddUser(user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to add user"})
		return
	}
	c.Redirect(http.StatusFound, "/users")
}

// editUserForm displays the form for editing an existing user.
func editUserForm(c *gin.Context) {
	username := c.Param("username")
	users, err := GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to load users"})
		return
	}

	var user *User
	for _, u := range users {
		if u.Username == username {
			user = &u
			break
		}
	}

	if user == nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"message": "User not found"})
		return
	}
	c.HTML(http.StatusOK, "user_form.tmpl", gin.H{"user": user, "action": fmt.Sprintf("/users/edit/%s", username)})
}

// updateUser handles updating an existing user.
func updateUser(c *gin.Context) {
	username := c.Param("username")
	users, err := GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to load users"})
		return
	}

	for i, u := range users {
		if u.Username == username {
			users[i].Admin = c.PostForm("admin") == "on"
			users[i].SuperAdmin = c.PostForm("super_admin") == "on"
			if password := c.PostForm("password"); password != "" {
				users[i].Password = hashPassword(password)
			}
			if err := SaveAllUsers(users); err != nil {
				c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to update user"})
				return
			}
			c.Redirect(http.StatusFound, "/users")
			return
		}
	}
	c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"message": "User not found"})
}

// deleteUser handles deleting a user.
func deleteUser(c *gin.Context) {
	username := c.Param("username")
	err := DeleteUser(username)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to delete user"})
		return
	}
	c.Redirect(http.StatusFound, "/users")
}

// newAttendanceForm displays the form for adding attendance.
func newAttendanceForm(c *gin.Context) {
	session := sessions.Default(c)
	username, ok := session.Get("username").(string)
	if !ok || username == "" {
		username = "Guest" // Fallback if username is missing
	}

	// Retrieve and clear the flash message if it exists
	flashMessage, _ := session.Get("flash").(string)
	session.Delete("flash") // Clear flash message after retrieval
	session.Save()

	// Get the current time in the correct format for datetime-local
	now := time.Now().Format("2006-01-02T15:04")

	// Pass the flash message and default datetime to the template
	c.HTML(http.StatusOK, "attendance_form.tmpl", gin.H{
		"action":   "/attendance/new",
		"username": username,
		"message":  flashMessage, // Display flash message if available
		"datetime": now,          // Pass current date and time
	})
}

func addAttendance(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")

	// Check if user is authenticated
	if username == nil {
		c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "User not authenticated"})
		return
	}

	// Use the username in the attendance record
	record := Attendance{
		Datetime:     c.PostForm("datetime"),
		FirstName:    c.PostForm("first_name"),
		LastName:     c.PostForm("last_name"),
		Reason:       c.PostForm("reason"),
		LicenceState: c.PostForm("licence_state"),
		User:         username.(string),
	}

	err := AddAttendance(record)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to add attendance"})
		return
	}

	// Set a flash message in the session
	session.Set("flash", "New record was added")
	session.Save()

	// Redirect to the form with the message
	c.Redirect(http.StatusFound, "/attendance/new")
}

// downloadAttendance exports all attendance records as JSON.
func downloadAttendance(c *gin.Context) {
	attendance, err := GetAllAttendance()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to load attendance"})
		return
	}

	// Set headers to download as a CSV file
	c.Header("Content-Disposition", "attachment; filename=attendance.csv")
	c.Header("Content-Type", "text/csv")

	// Create a CSV writer
	writer := csv.NewWriter(c.Writer)

	// Write CSV header
	writer.Write([]string{"Date", "First Name", "Last Name", "Reason", "Licence State", "Added By"})

	// Write attendance records to CSV
	for _, record := range attendance {
		writer.Write([]string{
			record.Datetime,
			record.FirstName,
			record.LastName,
			record.Reason,
			record.LicenceState,
			record.User,
		})
	}

	// Flush the writer to ensure all data is written to the response
	writer.Flush()

	// Check for any errors in writing
	if err := writer.Error(); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to write CSV"})
	}
}

// filterAttendance filters attendance records based on given parameters.
func filterAttendance(records []Attendance, date, firstName, lastName, reason, licenceState, user string) []Attendance {
	var filtered []Attendance
	for _, record := range records {
		recordDate := record.Datetime[:10] // Extract only the date part (YYYY-MM-DD)
		if (date == "" || recordDate == date) &&
			(firstName == "" || record.FirstName == firstName) &&
			(lastName == "" || record.LastName == lastName) &&
			(reason == "" || record.Reason == reason) &&
			(licenceState == "" || record.LicenceState == licenceState) &&
			(user == "" || record.User == user) {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// listAttendance displays attendance records with filter options.
func listAttendance(c *gin.Context) {
	// Retrieve the username from the session
	session := sessions.Default(c)
	username, ok := session.Get("username").(string)
	if !ok || username == "" {
		username = "Guest" // Default value if no user is logged in
	}

	// Load all users from the JSON file
	users, err := GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error retrieving user data"})
		return
	}

	// Determine the user's admin and superAdmin status
	var isAdmin, isSuperAdmin bool
	for _, user := range users {
		if user.Username == username {
			isAdmin = user.Admin
			isSuperAdmin = user.SuperAdmin
			break
		}
	}

	// Retrieve filter values from query parameters
	date := c.Query("date")
	firstName := c.Query("firstName")
	lastName := c.Query("lastName")
	reason := c.Query("reason")
	licenceState := c.Query("licence_state")
	filterUser := c.Query("user")

	// Retrieve all attendance records
	attendance, err := GetAllAttendance()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to load attendance"})
		return
	}

	// Filter the attendance records based on the parameters
	filteredAttendance := filterAttendance(attendance, date, firstName, lastName, reason, licenceState, filterUser)

	// Pass data to the template, including the filters and filtered results
	c.HTML(http.StatusOK, "attendance.tmpl", gin.H{
		"username":             username,
		"attendance":           filteredAttendance,
		"admin":                isAdmin,
		"superAdmin":           isSuperAdmin,
		"selectedDate":         date,
		"selectedUser":         filterUser,
		"selectedFirstName":    firstName,
		"selectedLastName":     lastName,
		"selectedReason":       reason,
		"selectedLicenceState": licenceState,
	})
}

// deleteAllAttendance deletes all attendance records.
// Only accessible to users with superAdmin privileges.
func deleteAllAttendance(c *gin.Context) {
	// Retrieve the username from the session
	session := sessions.Default(c)
	username, ok := session.Get("username").(string)
	if !ok || username == "" {
		// No username in the session, deny access
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{
			"message": "Access denied: insufficient privileges to delete all records",
		})
		return
	}

	// Load all users from the JSON file
	users, err := GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"message": "Error retrieving user data",
		})
		return
	}

	// Find the user in the list and check if they have superAdmin privileges
	var isSuperAdmin bool
	for _, user := range users {
		if user.Username == username {
			isSuperAdmin = user.SuperAdmin
			break
		}
	}

	// Check if the user has superAdmin privileges
	if !isSuperAdmin {
		// User does not have superAdmin privileges
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{
			"message": "Access denied: insufficient privileges to delete all records",
		})
		return
	}

	// Attempt to delete all attendance records by saving an empty list
	err = SaveAllAttendance([]Attendance{})
	if err != nil {
		// If there's an error saving the empty list, show an internal server error
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"message": "Failed to delete all attendance records",
		})
		return
	}

	// Redirect to the attendance page with a success message
	c.Redirect(http.StatusFound, "/attendance")
}
