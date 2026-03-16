# Scholarship Portal

A web application for managing student scholarship profiles, applications, and administration.

## Features

- User registration and authentication (JWT)
- Update and view user profiles
- Manage student academic information and selected majors
- Admin and student roles
- RESTful API with Swagger documentation

## Tech Stack 

- Go (Gin framework)
- PostgreSQL
- SQLC for type-safe queries
- JWT authentication

## Getting Started

1. **Clone the repository:**

   ```bash
    git clone https://github.com/kheav-kienghok/scholarship_portal.git
    cd scholarship_portal
   ```

2. **Configure your database in `.env` or config file.**

3. **Run database migrations:**

   ```bash
   goose -dir ./migrations postgres "your-db-url" up
   ```

4. **Start the server:**

   ```bash
   air
   ```

5. **API Docs:**  
   Visit `http://localhost:8080/swagger/index.html` for Swagger UI.
