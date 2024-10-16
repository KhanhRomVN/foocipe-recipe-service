# Foocipe Recipe Service

Foocipe Recipe Service is a backend application for managing cooking recipes. This service provides APIs for creating, retrieving, updating, and deleting recipes, as well as searching and filtering recipes based on various criteria.

## Features

- CRUD operations for recipes
- Search recipes by ingredients, cuisine type, or dietary restrictions
- User authentication and authorization
- Recipe rating and commenting system
- Ingredient management
- Meal planning functionality

## Technologies Used

- Go (Golang) for backend development
- PostgreSQL for database
- RESTful API design
- Docker for containerization

## Getting Started

### Prerequisites

- Go 1.16 or higher
- PostgreSQL 12 or higher
- Docker (optional)

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/KhanhRomVN/foocipe-recipe-service.git
   ```

2. Navigate to the project directory:
   ```
   cd foocipe-recipe-service
   ```

3. Install dependencies:
   ```
   go mod download
   ```

4. Set up environment variables (create a `.env` file based on `.env.example`)

5. Run the application:
   ```
   go run cmd/main.go
   ```

### Docker

To run the application using Docker:

1. Build the Docker image:
   ```
   docker build -t foocipe-recipe-service .
   ```

2. Run the container:
   ```
   docker run -p 8080:8080 foocipe-recipe-service
   ```

## API Documentation

API documentation is available at `/api/docs` when running the application.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.

## Author

KhanhRomVN

- GitHub: [@KhanhRomVN](https://github.com/KhanhRomVN)
- Project Repository: [foocipe-recipe-service](https://github.com/KhanhRomVN/foocipe-recipe-service)

## Acknowledgments

- Thanks to all contributors who have helped with this project.
- Inspired by the love for good food and efficient recipe management.