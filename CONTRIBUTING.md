# Contributing to TravelMind

Thank you for your interest in contributing to TravelMind! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

Be respectful, inclusive, and professional in all interactions with the community.

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+
- kubectl (for Kubernetes deployment)
- Git

### Local Setup

1. Clone the repository
```bash
git clone https://github.com/udaykishore-resu/travelmind.git
cd travelmind
```

2. Set up environment
```bash
cp .env.example .env
# Edit .env with your configuration
make dev-setup
```

3. Start services
```bash
make dev-up
make run-services
```

4. Verify setup
```bash
curl http://localhost:8080/health
```

## Development Workflow

### Branch Naming
- Feature: `feature/short-description`
- Bug fix: `fix/issue-number`
- Documentation: `docs/description`
- Refactor: `refactor/component-name`

### Commit Message Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:
```
feat(booking): add AI-powered itinerary builder

Implement Claude API integration for complex booking scenario analysis.
Includes risk assessment and travel recommendation engine.

Fixes #123
```

### Code Style

#### Go
- Follow `gofmt` and `golangci-lint` rules
- Run before commit:
```bash
make fmt
make lint
```

- Use meaningful variable names
- Add comments for exported functions
- Write tests for all public functions

#### Minimum Coverage
- New code must have 80%+ test coverage
- Use `make coverage` to check

### Testing

Run all tests:
```bash
make test
```

Run specific tests:
```bash
cd backend && go test -v -run TestBookingCreation ./...
```

Run with coverage:
```bash
make coverage
```

## Creating Pull Requests

1. Create a feature branch
```bash
git checkout -b feature/my-feature
```

2. Make changes and commit
```bash
git add .
git commit -m "feat(service): description of changes"
```

3. Push branch
```bash
git push origin feature/my-feature
```

4. Open PR on GitHub with:
   - Clear title: `Add AI-powered booking recommendations`
   - Description: What changed and why
   - Link any related issues: `Fixes #123`
   - Checklist:
     - [ ] Tests added/updated
     - [ ] Documentation updated
     - [ ] Code reviewed locally
     - [ ] No linting errors

## Architecture Guidelines

### Service Design
- Each service should have a single responsibility
- Use dependency injection for testability
- Implement proper error handling with context
- Add structured logging for debugging

### Database
- Use transactions for multi-step operations (Saga pattern)
- Index foreign keys and commonly filtered columns
- Add migrations for schema changes
- Document complex queries

### API Design
- Use REST conventions for public APIs
- Return consistent error responses
- Document all endpoints with OpenAPI/Swagger
- Version APIs for backward compatibility

### Performance
- Cache frequently accessed data
- Use database connection pooling
- Implement circuit breakers for external calls
- Profile before optimizing

## Documentation

### Code Comments
```go
// ProcessBooking handles the complete booking workflow.
// It validates the itinerary, checks inventory, processes payment,
// and confirms with suppliers in sequence.
func (s *BookingService) ProcessBooking(ctx context.Context, booking *Booking) error {
```

### Markdown
- Use clear headings
- Include code examples
- Add diagrams for complex flows
- Keep lines under 80 characters

## Deployment

### Staging Deployment
```bash
make docker-build
make docker-push
make k8s-deploy
```

### Production Deployment
- Create a release branch
- Tag version: `git tag v1.0.0`
- Canary deployment with monitoring (20% traffic)
- Promote to 100% after verification

## Common Tasks

### Add a New Endpoint
1. Define request/response models in `internal/models`
2. Create handler in `internal/handlers`
3. Add route in `cmd/api-gateway/main.go`
4. Write tests in `internal/handlers/*_test.go`
5. Document in OpenAPI spec

### Add a Database Migration
1. Create file: `backend/migrations/NNN_description.sql`
2. Use provided schema and conventions
3. Test: `make db-migrate-up`
4. Test rollback: `make db-migrate-down`

### Add a New Service
1. Create directory: `backend/services/new-service`
2. Implement service interface
3. Add gRPC definition if service-to-service
4. Create Dockerfile
5. Add Kubernetes manifests

## Getting Help

- Check existing issues and documentation
- Ask in GitHub discussions
- Email: udaykishoreresu2@gmail.com

## Code Review Process

Reviewers will check:
- ✅ Correctness: Does it do what it says?
- ✅ Tests: Are there adequate tests?
- ✅ Performance: Any N+1 queries or memory leaks?
- ✅ Security: Any vulnerabilities or data exposure?
- ✅ Documentation: Is it clear and complete?
- ✅ Style: Does it follow conventions?

## Versioning

We follow [Semantic Versioning](https://semver.org/):
- MAJOR: Breaking API changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes

## Security

If you find a security vulnerability, please email udaykishoreresu2@gmail.com instead of using the issue tracker.

## License

By contributing to TravelMind, you agree that your contributions will be licensed under the same license as the project (MIT License).

---

Happy coding! 🚀
