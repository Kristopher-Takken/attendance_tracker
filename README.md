# Attendance Management Web Application Documentation

**MIT License**

```
MIT License

Copyright (c) 2023 Allform Software Solutions

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

*The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.*

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

# Attendance Management Web Application Documentation

This documentation provides a comprehensive guide to the Attendance Management Web Application, designed to assist different audiences:

- **Software Integrators (Server-Side)**: Setup, installation, and understanding of the application's architecture and codebase.
- **Administrators**: User and attendance management, including roles and permissions.
- **General Users**: Basic usage, such as logging in and adding attendance records.

---

## Table of Contents

- [Introduction](#introduction)
- [For Software Integrators](#for-software-integrators)
  - [Prerequisites](#prerequisites)
  - [Setup and Installation](#setup-and-installation)
  - [Application Architecture](#application-architecture)
  - [Code Overview](#code-overview)
  - [Security Considerations](#security-considerations)
- [For Administrators](#for-administrators)
  - [User Management](#user-management)
  - [Attendance Management](#attendance-management)
  - [Roles and Permissions](#roles-and-permissions)
- [For General Users](#for-general-users)
  - [Accessing the Application](#accessing-the-application)
  - [Authentication](#authentication)
  - [Adding Attendance Records](#adding-attendance-records)
  - [Viewing Attendance Records](#viewing-attendance-records)
- [Frequently Asked Questions](#frequently-asked-questions)

---

## Introduction

The **Attendance Management Web Application** is a web-based tool designed to streamline attendance tracking within an organization. It provides user authentication with role-based access control, allowing administrators to manage users and attendance records, while general users can log their attendance entries.

---

## For Software Integrators

This section provides detailed information for setting up and integrating the application on the server side.

### Prerequisites

Before installing the application, ensure the following prerequisites are met:

- **Go Programming Language**: Version 1.15 or newer.
  - Download and install from [golang.org/dl](https://golang.org/dl/).
- **Git**: For cloning repositories (if applicable).
- **Environment**: Access to a Unix-like terminal (Linux, macOS) or Windows Command Prompt/PowerShell.

### Setup and Installation

Follow these steps to set up the application:

#### 1. Obtain the Source Code

- **Option 1**: Clone from a repository (if hosted on GitHub or similar):

  ```bash
  git clone https://github.com/your-username/attendance-app.git
  cd attendance-app
  ```

- **Option 2**: Download the source code directly and extract it to a directory.

#### 2. Install Dependencies

The application uses the following third-party packages:

- `github.com/gin-gonic/gin`: Web framework.
- `github.com/gin-contrib/sessions`: Session management.
- `golang.org/x/crypto/bcrypt`: Password hashing.

Use Go Modules to manage dependencies:

```bash
go mod init attendance-app
go mod tidy
```

This will create a `go.mod` file and download necessary packages.

#### 3. File Structure

Ensure the following files and directories are present:

```
attendance-app/
├── main.go
├── models.go
├── database.go
├── templates/
│   ├── index.tmpl
│   ├── login.tmpl
│   ├── error.tmpl
│   ├── users.tmpl
│   ├── user_form.tmpl
│   ├── attendance.tmpl
│   └── attendance_form.tmpl
├── static/
│   └── [static files like CSS, JS]
├── users.json
└── attendance.json
```

- **main.go**: Entry point of the application.
- **models.go**: Data models for users and attendance.
- **database.go**: Functions for data persistence.
- **templates/**: HTML templates for rendering pages.
- **static/**: Static assets (CSS, JS).
- **users.json** & **attendance.json**: Data storage files (will be created if they don't exist).

#### 4. Build and Run the Application

Compile and run the application:

```bash
go run main.go models.go database.go
```

Alternatively, build an executable:

```bash
go build -o attendance-app
./attendance-app
```

The server will start and listen on port **80** by default.

#### 5. Access the Application

Open a web browser and navigate to:

```
http://localhost/
```

### Application Architecture

The application follows the Model-View-Controller (MVC) pattern using the Gin web framework.

- **Models**: Defined in `models.go` (User, Attendance).
- **Views**: HTML templates in the `templates/` directory.
- **Controllers**: Route handlers in `main.go`.

**Data Persistence**:

- Data is stored in JSON files (`users.json`, `attendance.json`).
- Mutex (`sync.Mutex`) is used for thread-safe read/write operations.

**Sessions**:

- Managed using cookie-based sessions (`gin-contrib/sessions`).

### Code Overview

#### main.go

- **Initialization**:
  - Checks for existing users.
  - Creates a default superadmin user (`admin`/`admin`) if none exist.

- **Router Setup**:
  - Loads HTML templates and static files.
  - Configures session middleware.

- **Routes**:
  - **Authentication**:
    - `GET /login`: shows the login form.
    - `POST /login`: handles user login.
    - `GET /logout`: logs out the user.
  - **User Management** (requires admin):
    - `GET /users`: lists users.
    - `GET /users/new`: form to add a new user.
    - `POST /users/new`: adds a new user.
    - `GET /users/edit/:username`: form to edit a user.
    - `POST /users/edit/:username`: updates a user.
    - `POST /users/delete/:username`: deletes a user.
  - **Attendance Management**:
    - `GET /attendance`: lists attendance records.
    - `GET /attendance/new`: form to add attendance.
    - `POST /attendance/new`: adds attendance record.
    - `GET /attendance/download` (requires admin): downloads attendance as CSV.
    - `POST /attendance/delete_all` (requires superadmin): deletes all attendance records.

- **Middleware**:
  - `requireLoginMiddleware`: ensures user is logged in.
  - `requireAdmin`: restricts access to admins.
  - `requireSuperAdmin`: restricts access to superadmins.

#### models.go

Defines the data models:

- **User**:
  - `Username`
  - `Password` (hashed)
  - `Admin` (boolean)
  - `SuperAdmin` (boolean)
  
- **Attendance**:
  - `Datetime`
  - `FirstName`
  - `LastName`
  - `Reason`
  - `LicenceState`
  - `User` (who added the record)

#### database.go

Handles data persistence:

- **User Functions**:
  - `GetAllUsers()`
  - `SaveAllUsers(users []User)`
  - `AddUser(user User)`
  - `UpdateUser(updated User)`
  - `DeleteUser(username string)`

- **Attendance Functions**:
  - `GetAllAttendance()`
  - `SaveAllAttendance(attendance []Attendance)`
  - `AddAttendance(record Attendance)`

- **Password Functions**:
  - `hashPassword(password string)`
  - `verifyPassword(hashedPassword, password string)`

### Security Considerations

- **Password Hashing**:
  - Passwords are hashed using bcrypt before storage.

- **Session Security**:
  - Sessions are managed using secure cookies.
  - Ensure the session secret (`[]byte("secret")`) is replaced with a strong, unique key.

- **Access Control**:
  - Middleware functions enforce role-based access control.

- **Data Storage**:
  - Consider moving from JSON files to a database (e.g., MySQL, PostgreSQL) for production use.

- **Port Usage**:
  - The application runs on port **80**. Ensure this port is open and not in use by another service.

---

## For Administrators

As an administrator, you have elevated permissions to manage users and attendance records.

### User Management

#### Accessing User Management

- **Navigate to** `/users` after logging in as an admin.

#### Adding Users

1. **Go to** `/users/new`.
2. **Fill Out the Form**:
   - **Username**: Unique username.
   - **Password**: Secure password.
   - **Admin**: Check if the user should have admin privileges.
   - **SuperAdmin**: Check if the user should have superadmin privileges.
3. **Submit**: Click "Add User".

#### Editing Users

1. **Go to** `/users`.
2. **Click** "Edit" next to the user.
3. **Update Details** as needed.
4. **Submit**: Click "Update User".

#### Deleting Users

1. **Go to** `/users`.
2. **Click** "Delete" next to the user.
3. **Confirm Deletion** if prompted.

### Attendance Management

#### Viewing Attendance Records

- **Navigate to** `/attendance`.
- **Filters**:
  - Use the provided filters (Date, First Name, Last Name, Reason, Licence State, User) to narrow down records.

#### Downloading Attendance Records

- **Navigate to** `/attendance/download` to download all attendance records as a CSV file.

#### Deleting All Attendance Records (SuperAdmin Only)

- **Navigate to** `/attendance`.
- **Click** "Delete All Records" (visible only to superadmins).
- **Confirm Deletion**.

**Warning**: This action is irreversible.

### Roles and Permissions

#### Admin

- Can manage users and attendance records.
- Access to user management pages.

#### SuperAdmin

- All permissions of Admin.
- Additionally, can delete all attendance records.
- The initial `admin` user created by default is a superadmin.

---

## For General Users

As a general user, you can log in to the application and add attendance records.

### Accessing the Application

- Open a web browser and navigate to the application URL (e.g., `http://localhost/`).

### Authentication

#### Logging In

1. **Go to** `/login`.
2. **Enter** your username and password.
3. **Click** "Login".

#### Logging Out

- **Click** "Logout" to end your session.

### Adding Attendance Records

1. **Navigate to** `/attendance/new`.
2. **Fill Out the Form**:
   - **Date and Time**: Defaults to the current date and time.
   - **First Name**: Your first name.
   - **Last Name**: Your last name.
   - **Reason**: Select or enter the reason for attendance.
   - **Licence State**: Enter your licence state.
3. **Submit**: Click "Add Attendance".

### Viewing Attendance Records

- **Navigate to** `/attendance` to see all attendance records.
- **Filtering**:
  - Use the filters to search for specific records.

---

## Frequently Asked Questions

**Q1: What are the default login credentials?**

- **A**: If no users exist, a default superadmin is created:
  - **Username**: `admin`
  - **Password**: `admin`
- **Note**: It is highly recommended to change this password after the first login.

**Q2: How do I change the session secret key?**

- **A**: In `main.go`, replace `[]byte("secret")` with a unique, secure byte slice.

  ```go
  store := cookie.NewStore([]byte("your-secure-random-key"))
  ```

- **Tip**: Use a cryptographically secure method to generate the key.

**Q3: Can I run the application on a different port?**

- **A**: Yes. In `main.go`, modify the `router.Run(":80")` line, replacing `80` with your desired port.

**Q4: How do I secure the application for production use?**

- **A**:
  - Use HTTPS by configuring TLS certificates.
  - Replace JSON file storage with a secure database.
  - Implement proper input validation and sanitization.
  - Regularly update dependencies to patch security vulnerabilities.

**Q5: What should I do if I forget my admin password?**

- **A**: Manually edit the `users.json` file or recreate it:
  - Stop the application.
  - Delete `users.json` (if you're okay with recreating users).
  - Restart the application to generate a new default admin user.

**Q6: How can I back up attendance records?**

- **A**: Regularly download attendance records via `/attendance/download` and back up the `attendance.json` file.

**Q7: Can I customize the reasons for attendance?**

- **A**: Modify the template `attendance_form.tmpl` to include predefined reasons or allow custom input.

**Q8: How do I add new fields to the attendance records?**

- **A**:
  - Update the `Attendance` struct in `models.go`.
  - Modify forms and templates to include the new fields.
  - Update data handling functions in `database.go`.

---

For further assistance, refer to the source code or contact the development team.

---

# End of Documentation

This documentation covers the essential aspects of the application for different audiences, providing guidance on setup, usage, and administration. It should serve as a comprehensive resource for integrating and using the Attendance Management Web Application.
