# bc-go

A Go client library for the Microsoft Dynamics 365 Business Central API.

## Project Overview

This is a Go SDK for interacting with the Microsoft Dynamics 365 Business Central OData API. The library provides a clean, idiomatic Go interface for making requests to both the standard v2.0 API endpoints and custom extension APIs.

## Design Philosophy

### Modern Go SDK Conventions
- Follow standard Go idioms and patterns
- Use options structs for configuration
- Provide clear, type-safe interfaces
- Include comprehensive error handling
- Support standard library interfaces where applicable
- Use context for cancellation and timeouts

### Balance Power with Simplicity
As an OData library, we aim to:
- **Power & Flexibility**: Support the full range of OData query capabilities (filtering, expansion, selection, ordering, etc.)
- **Simplicity**: Keep the API surface intuitive and not overly complex
- **Avoid Over-Engineering**: Don't try to abstract away OData completely - users should understand they're working with OData
- **Practical Defaults**: Provide sensible defaults while allowing customization when needed

### Key Principles
1. **Don't hide OData**: Users should be comfortable with OData concepts
2. **Type safety where it matters**: Use strong types for configuration and responses
3. **Flexibility over magic**: Prefer explicit configuration over implicit behavior
4. **Progressive disclosure**: Simple tasks should be simple, complex tasks should be possible

## Architecture

- **Client**: One client per Route (publisher/group/version) and CompanyID combination
- **Authentication**: Uses Azure AD (Entra ID) token-based authentication via `azcore.TokenCredential`
- **HTTP**: Built on standard `net/http` with configurable timeouts and clients
- **Logging**: Uses structured logging via `log/slog`

## Module Versioning

Currently in v0.x to allow API iteration. Will move to v1.0.0 when the API stabilizes and we're ready to commit to backward compatibility.
