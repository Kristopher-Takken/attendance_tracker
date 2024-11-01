// database.go
package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	usersFile      = "users.json"
	attendanceFile = "attendance.json"
	mu             sync.Mutex // For thread-safe operations
)

// ---- User Management Functions ----

// GetAllUsers retrieves all users from the users.json file.
func GetAllUsers() ([]User, error) {
	mu.Lock()
	defer mu.Unlock()

	var users []User
	file, err := ioutil.ReadFile(usersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return users, nil // No users file yet, return empty list
		}
		return nil, err
	}
	err = json.Unmarshal(file, &users)
	return users, err
}

// SaveAllUsers saves all users to the users.json file.
func SaveAllUsers(users []User) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(usersFile, data, 0644)
}

// AddUser adds a new user to the users.json file.
func AddUser(user User) error {
	users, err := GetAllUsers()
	if err != nil {
		return err
	}
	users = append(users, user)
	return SaveAllUsers(users)
}

// UpdateUser updates an existing user in the users.json file.
func UpdateUser(updated User) error {
	users, err := GetAllUsers()
	if err != nil {
		return err
	}
	for i, user := range users {
		if user.Username == updated.Username {
			users[i] = updated
			return SaveAllUsers(users)
		}
	}
	return errors.New("user not found")
}

// DeleteUser removes a user from the users.json file by username.
func DeleteUser(username string) error {
	users, err := GetAllUsers()
	if err != nil {
		return err
	}
	for i, user := range users {
		if user.Username == username {
			users = append(users[:i], users[i+1:]...)
			return SaveAllUsers(users)
		}
	}
	return errors.New("user not found")
}

// ---- Password Hashing Functions ----

// hashPassword hashes a plain text password.
func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

// verifyPassword checks if the provided password matches the hashed password.
func verifyPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// ---- Attendance Management Functions ----

// GetAllAttendance retrieves all attendance records from the attendance.json file.
func GetAllAttendance() ([]Attendance, error) {
	mu.Lock()
	defer mu.Unlock()

	var attendance []Attendance
	file, err := ioutil.ReadFile(attendanceFile)
	if err != nil {
		if os.IsNotExist(err) {
			return attendance, nil // No attendance file yet, return empty list
		}
		return nil, err
	}
	err = json.Unmarshal(file, &attendance)
	return attendance, err
}

// SaveAllAttendance saves all attendance records to the attendance.json file.
func SaveAllAttendance(attendance []Attendance) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(attendance, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(attendanceFile, data, 0644)
}

// AddAttendance adds a new attendance record to the attendance.json file.
func AddAttendance(record Attendance) error {
	attendance, err := GetAllAttendance()
	if err != nil {
		return err
	}
	attendance = append(attendance, record)
	return SaveAllAttendance(attendance)
}
