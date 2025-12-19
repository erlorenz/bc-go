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

### HTTP Client and Transport Handling

The library respects custom HTTP clients provided by users (e.g., for OpenTelemetry instrumentation) while adding Business Central-specific functionality:

1. **Transport Extraction**: When a user provides an `http.Client`, the library extracts its `Transport`
2. **Transport Wrapping**: The extracted transport is wrapped in `bcTransport`, which adds:
   - Authentication (Bearer token via Azure AD)
   - User-Agent header
   - Accept header (application/json with no OData metadata)
3. **Client Reconstruction**: A new `http.Client` is created with the wrapped transport, preserving the original client's:
   - Timeout settings
   - Cookie jar
   - Redirect policy
   - CheckRedirect function

**Request flow**: `Your code → bcTransport (auth/headers) → User's transport (e.g., OTel) → Network`

This design ensures that instrumentation, proxies, and other transport-level customizations continue to work while the library handles Business Central authentication transparently.

## Module Versioning

Currently in v0.x to allow API iteration. Will move to v1.0.0 when the API stabilizes and we're ready to commit to backward compatibility.

## Business Central API Behavior Notes

### OData Query Options with POST/PATCH

**Important**: Business Central has specific behavior with OData query options on POST and PATCH requests:

- **$select works**: You can use `$select` in POST/PATCH requests and the response will only include the selected fields
- **$expand does NOT work**: Including `$expand` in POST/PATCH requests has no effect - the response will only contain the base entity without any expanded related entities
- **Prefer header has no effect**: Setting `Prefer: return=representation` or `Prefer: return=minimal` does not change the response behavior

### Implementation Impact

Because of this limitation, the Create and Update methods in APIPage use a smart two-step approach when expands are needed:

1. **No expands**: Single API call (POST/PATCH) - returns base entity with selects applied
2. **With expands**: Two API calls:
   - First: POST/PATCH to create/update the entity (selects applied here)
   - Second: GET with expands from live database to retrieve the entity with expanded related entities

This ensures that expands work correctly while minimizing API calls when expands aren't needed.

**Note**: The DefaultExpand field on APIPage is automatically included when checking if a second GET call is needed.

## OData Metadata and ETags

### Metadata Level

The library uses `Accept: application/json;odata.metadata=minimal` which provides a balance between clean responses and useful metadata:

**Includes in responses**:
- `@odata.etag` - Entity version tag for optimistic concurrency control
- `@odata.context` - Metadata context URL

**Excludes**:
- Full type annotations and other verbose metadata

### ETags and Optimistic Concurrency

**ETags** prevent lost updates in multi-user scenarios by detecting concurrent modifications:

1. **GET** a record → Response body includes `@odata.etag` field (e.g., `"W/\"JzQ0O0...==\""`)
2. **UPDATE** with that ETag → Include it in `UpdateOptions.ETag`
3. **Server validates**:
   - ✅ ETag matches → Update succeeds
   - ❌ ETag changed → Returns `409 Conflict` with error code `Request_EntityChanged`

**Example usage**:

```go
// Read-Modify-Write with ETag protection
type Customer struct {
    ID    uuid.UUID `json:"id"`
    Name  string    `json:"name"`
    ETag  string    `json:"@odata.etag"`
}

customer, err := api.GetByID(ctx, id, nil)
// customer.ETag contains the version tag

// Modify
customer.Name = "New Name"

// Update with ETag protection
updated, err := api.Update(ctx, id, customer, &UpdateOptions{
    ETag: customer.ETag,
})
// Returns error if someone else modified the record
```

**Error handling**:

```go
err := api.Delete(ctx, id, etag)
if err != nil {
    // Check for concurrency conflict
    if strings.Contains(err.Error(), "Request_EntityChanged") {
        // Handle: record was modified by another user
        // Typical response: 409 Conflict with error message
    }
}
```

**Default behavior**: If no ETag is provided, the library uses `If-Match: *` which means "update regardless of current state" (no concurrency protection).

## Pagination and Iterators

### Iterator Pattern for List Operations

The library uses Go 1.23+ `iter.Seq2[T, error]` for list operations, following modern SDK patterns (similar to Stripe Go SDK v82+):

```go
// List returns an iterator that handles pagination automatically
for customer, err := range api.List(ctx, &ListOptions{Filter: "status eq 'active'"}) {
    if err != nil {
        return err  // Pagination error
    }
    process(customer)  // Process each item as it arrives
}
```

**Benefits**:
- **Memory efficient**: Items are processed as they arrive, not all loaded at once
- **Lazy evaluation**: Stops fetching when iteration stops (early exit)
- **Simple API**: Auto-pagination is transparent to the user
- **Idiomatic Go**: Matches the `for...range` pattern from Go 1.23+

**Common patterns**:

```go
// Collect into slice
var customers []Customer
for customer, err := range api.List(ctx, nil) {
    if err != nil {
        return err
    }
    customers = append(customers, customer)
}

// Build map directly (no intermediate slice)
customerMap := make(map[uuid.UUID]Customer)
for customer, err := range api.List(ctx, nil) {
    if err != nil {
        return err
    }
    customerMap[customer.GetID()] = customer
}

// Process at ACL boundary (typical usage)
for customer, err := range api.List(ctx, nil) {
    if err != nil {
        return err
    }
    domainCustomer := adapter.ToDomain(customer)
    if err := domainCustomer.Validate(); err != nil {
        log.Warn("skipping invalid customer", "id", customer.GetID())
        continue
    }
    results = append(results, domainCustomer)
}
```

### Server-Side Paging Control

Business Central's default page size is **20,000 items**, which can be excessive for async processing or memory-constrained environments. Use `MaxPageSize` to control server-side pagination:

```go
opts := &ListOptions{
    MaxPageSize: 100,  // Fetch 100 items per page from server
    Filter: "status eq 'active'",
}

for customer, err := range api.List(ctx, opts) {
    if err != nil {
        return err
    }
    process(customer)
}
```

This sets the `Prefer: odata.maxpagesize=100` header, instructing Business Central to return smaller pages.

### Opaque nextLink Handling

Per the OData specification, `@odata.nextLink` URLs are **opaque** - they should be used as-is without parsing or modification.

**How it works**:
1. First page: Library builds URL from entity set name + query parameters
2. Subsequent pages: Library uses the full `@odata.nextLink` URL exactly as BC provides it
3. BC includes all necessary state in the nextLink (`$skiptoken`, `aid`, filters, etc.)

**Implementation**:
- `DoRequest` detects full URLs (http/https prefix) and uses them directly
- Query encoding is skipped for opaque URLs (parameters already in URL)
- No parsing, no assumptions about URL structure

This matches the pattern used by Microsoft Graph Go SDK and ASP.NET OData client libraries.

**Why this matters**:
- ✅ Preserves all BC pagination state (`$skiptoken`, custom params like `aid`)
- ✅ Forward-compatible if BC changes pagination mechanism
- ✅ Follows OData best practices
- ✅ Simpler implementation (no URL parsing/reconstruction)
