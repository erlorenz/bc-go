package bc

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"slices"

	"github.com/google/uuid"
)

type Record interface {
	GetID() uuid.UUID
}

// ODataETag provides the @odata.etag field for optimistic concurrency control.
// Embed this in your record types to automatically capture ETags from API responses.
//
// Example:
//
//	type Customer struct {
//	    ID   uuid.UUID `json:"id"`
//	    Name string    `json:"displayName"`
//	    bc.ODataETag
//	}
//
//	// Get entity with ETag
//	customer, _ := api.GetByID(ctx, id, nil)
//
//	// Modify and update with concurrency protection
//	customer.Name = "Updated Name"
//	updated, err := api.Update(ctx, id, customer, &UpdateOptions{
//	    ETag: customer.ETag,  // Use captured ETag
//	})
//	if err != nil {
//	    // Handle 409 Conflict if record was modified by another user
//	}
type ODataETag struct {
	ETag string `json:"@odata.etag"`
}

// APIPage represents an API page in Business Central.
// It has the CRUD methods as well as a List method that returns
// a list of records[T].
//
// Use DefaultFilter and DefaultExpand to set query parameters that will be
// applied to all requests for this record type.
type APIPage[T Record] struct {
	entitySetName string
	client        *Client

	// DefaultFilter is applied to all queries (e.g., "status eq 'active'").
	// Merged with per-request filters using AND logic.
	DefaultFilter string

	// DefaultExpand is merged with per-request expands.
	// Useful for always including certain related entities (e.g., []string{"addresses"}).
	DefaultExpand []string
}

// NewAPIPage creates an instance of an APIPage. It panics if client or entitySetName are empty.
//
// Example:
//
//	api := bc.NewAPIPage[Customer](client, "customers")
//	api.DefaultFilter = "status eq 'active'"
//	api.DefaultExpand = []string{"addresses", "contacts"}
func NewAPIPage[T Record](client *Client, entitySetName string) *APIPage[T] {
	if client == nil {
		panic("create API page: client is nil")
	}

	if entitySetName == "" {
		panic("create API page: entitySetName is empty")
	}
	return &APIPage[T]{
		entitySetName: entitySetName,
		client:        client,
	}
}

type GetOptions struct {
	Select    []string
	Expand    []string
	UseLiveDB bool
}

type ListOptions struct {
	Filter      string   // OData $filter expression
	OrderBy     string   // OData $orderby expression
	Select      []string // Fields to select
	Expand      []string // Related entities to expand
	Top         int      // Limit total results ($top parameter, 0 = no limit)
	MaxPageSize int      // Server-side page size (Prefer: odata.maxpagesize, 0 = BC default of 20,000)
	UseLiveDB   bool     // Use live database instead of read replica
}

// GetByID makes a GET request to the endpoint and retrieves a single record T.
// Requires the ID and takes an optional slice of expand strings.
// DefaultExpand is automatically merged with any provided expands.
func (a *APIPage[T]) GetByID(ctx context.Context, id uuid.UUID, opts *GetOptions) (T, error) {
	path := a.pathWithID(id)

	query := &ODataQuery{}

	if opts != nil {
		query.Select = opts.Select
		query.Expand = opts.Expand
	}

	// Merge in DefaultExpand
	query.Expand = a.mergeExpands(query.Expand)

	// Default to using read replica
	var options []RequestOption
	if opts == nil || !opts.UseLiveDB {
		options = append(options, WithDataAccessReadOnly())
	}

	a.client.logger.DebugContext(ctx, "Making GetByID request.",
		"entitySetName", a.entitySetName, "id", id, "query", query, "readReplica", len(options) > 0)

	resp, err := a.client.DoRequest(ctx, http.MethodGet, path, query, nil, options...)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("GetByID request with id %s: %w", id, err)
	}

	var entity T
	if err := resp.DecodeJSON(&entity); err != nil {
		var zero T
		return zero, fmt.Errorf("decoding %s with id %s: %w", a.entitySetName, id, err)
	}

	a.client.logger.DebugContext(ctx, "Successfully retrieved record.", "entitySetName", a.entitySetName, "id", id)
	return entity, nil
}

// List returns an iterator that fetches entities page by page using OData pagination.
// Auto-pagination is handled automatically via @odata.nextLink.
//
// Example - collect all items:
//
//	var customers []Customer
//	for customer, err := range api.List(ctx, nil) {
//	    if err != nil {
//	        return err
//	    }
//	    customers = append(customers, customer)
//	}
//
// Example - process directly (more memory efficient):
//
//	for customer, err := range api.List(ctx, &ListOptions{Filter: "status eq 'active'"}) {
//	    if err != nil {
//	        log.Error("pagination failed", "error", err)
//	        break
//	    }
//	    domainCustomer := adapter.ToDomain(customer)
//	    process(domainCustomer)
//	}
//
// Example - build map directly:
//
//	customerMap := make(map[uuid.UUID]Customer)
//	for customer, err := range api.List(ctx, nil) {
//	    if err != nil {
//	        return err
//	    }
//	    customerMap[customer.GetID()] = customer
//	}
//
// Server-side paging (RECOMMENDED for large datasets):
// Use MaxPageSize to control page size. BC default is 20,000 which can be excessive
// for async processing or memory-constrained environments.
//
//	opts := &ListOptions{
//	    MaxPageSize: 100,  // Fetch 100 items per page from server
//	    Filter: "status eq 'active'",
//	}
//	for customer, err := range api.List(ctx, opts) {
//	    if err != nil {
//	        return err
//	    }
//	    process(customer)
//	}
func (a *APIPage[T]) List(ctx context.Context, opts *ListOptions) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		// Build query from options
		query := &ODataQuery{}
		if opts != nil {
			query.Filter = a.mergeFilter(opts.Filter)
			query.OrderBy = opts.OrderBy
			query.Select = opts.Select
			query.Expand = a.mergeExpands(opts.Expand)
			query.Top = opts.Top
		} else {
			query.Filter = a.DefaultFilter
			query.Expand = a.DefaultExpand
		}

		// Build request options
		var reqOpts []RequestOption
		if opts == nil || !opts.UseLiveDB {
			reqOpts = append(reqOpts, WithDataAccessReadOnly())
		}

		// Set server-side page size if specified
		if opts != nil && opts.MaxPageSize > 0 {
			reqOpts = append(reqOpts, WithMaxPageSize(opts.MaxPageSize))
		}

		// Track pagination state
		nextLink := ""
		pageNum := 0
		totalItems := 0

		for {
			pageNum++

			// Determine what to fetch
			var path string
			var currentQuery *ODataQuery

			if nextLink != "" {
				// BC provides full @odata.nextLink URL - use as opaque per OData spec
				// DoRequest will detect the http:// prefix and use it as-is
				path = nextLink
				currentQuery = nil // Ignored for opaque URLs
			} else {
				// First page - use entity set name + query
				path = a.entitySetName
				currentQuery = query
			}

			a.client.logger.DebugContext(ctx, "Fetching page",
				"entitySetName", a.entitySetName,
				"page", pageNum,
				"totalSoFar", totalItems)

			// Fetch page
			resp, err := a.client.DoRequest(ctx, http.MethodGet, path, currentQuery, nil, reqOpts...)
			if err != nil {
				var zero T
				yield(zero, fmt.Errorf("fetching page %d: %w", pageNum, err))
				return
			}

			// Decode response - BC returns {"value": [...], "@odata.nextLink": "..."}
			var listResp struct {
				Value    []T    `json:"value"`
				NextLink string `json:"@odata.nextLink"`
			}

			if err := resp.DecodeJSON(&listResp); err != nil {
				var zero T
				yield(zero, fmt.Errorf("decoding page %d: %w", pageNum, err))
				return
			}

			// Yield each item
			for _, item := range listResp.Value {
				totalItems++
				if !yield(item, nil) {
					// Consumer stopped iteration
					a.client.logger.DebugContext(ctx, "Iteration stopped by consumer",
						"entitySetName", a.entitySetName,
						"totalItems", totalItems,
						"pages", pageNum)
					return
				}
			}

			// Check for next page
			if listResp.NextLink == "" {
				// No more pages
				a.client.logger.DebugContext(ctx, "Completed pagination",
					"entitySetName", a.entitySetName,
					"totalItems", totalItems,
					"pages", pageNum)
				return
			}

			nextLink = listResp.NextLink
		}
	}
}

type UpdateOptions struct {
	Expand []string
	Select []string
	ETag   string // Optional ETag for optimistic concurrency. If empty, uses If-Match: *
}

// Update makes a PATCH request to update the entity and returns the updated entity T.
// Behavior depends on expands:
//   - No expands (DefaultExpand or opts.Expand): Returns entity from PATCH response (1 API call)
//   - With expands: Does PATCH + GET to retrieve expands correctly (2 API calls)
//
// Note: Business Central PATCH responses don't include expanded entities, only the base entity.
// Selects work in the PATCH response, but expands require a separate GET call.
func (a *APIPage[T]) Update(ctx context.Context, id uuid.UUID, body any, opts *UpdateOptions) (T, error) {

	a.client.logger.DebugContext(ctx, "Making Update request.", "entitySetName", a.entitySetName, "id", id, "body", body)

	// Build query for selects (if provided)
	query := &ODataQuery{}
	if opts != nil {
		query.Select = opts.Select
	}

	// Build request options for ETag (if provided)
	var reqOpts []RequestOption
	if opts != nil && opts.ETag != "" {
		reqOpts = append(reqOpts, WithETag(opts.ETag))
	}

	// Make PATCH request
	resp, err := a.client.DoRequest(ctx, http.MethodPatch, a.pathWithID(id), query, body, reqOpts...)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("PATCH request with id %s: %w", id, err)
	}

	// Check if there are ANY expands (DefaultExpand or opts.Expand)
	hasExpands := len(a.DefaultExpand) > 0 || (opts != nil && len(opts.Expand) > 0)

	// If no expands, decode and return the PATCH response directly
	if !hasExpands {
		var entity T
		if err := resp.DecodeJSON(&entity); err != nil {
			var zero T
			return zero, fmt.Errorf("decoding updated entity: %w", err)
		}
		a.client.logger.Debug("Successfully updated record.", "entitySetName", a.entitySetName, "id", id)
		return entity, nil
	}

	// If expands needed, make GET request to retrieve entity with expands from live DB
	a.client.logger.Debug("Successfully updated record, retrieving with expands.", "entitySetName", a.entitySetName, "id", id)
	return a.GetByID(ctx, id, &GetOptions{
		Expand:    opts.Expand,
		Select:    opts.Select,
		UseLiveDB: true,
	})
}

type CreateOptions struct {
	Expand []string
	Select []string
}

// Create makes a POST request to create an entity and returns the created entity T.
// Behavior depends on expands:
//   - No expands (DefaultExpand or opts.Expand): Returns entity from POST response (1 API call)
//   - With expands: Does POST + GET to retrieve expands correctly (2 API calls)
//
// Note: Business Central POST responses don't include expanded entities, only the base entity.
// Selects work in the POST response, but expands require a separate GET call.
func (a *APIPage[T]) Create(ctx context.Context, body any, opts *CreateOptions) (T, error) {
	// Build query for selects (if provided)
	query := &ODataQuery{}
	if opts != nil {
		query.Select = opts.Select
	}

	a.client.logger.DebugContext(ctx, "Making Create request.", "entitySetName", a.entitySetName, "body", body, "query", query)

	// Make POST request
	resp, err := a.client.DoRequest(ctx, http.MethodPost, a.entitySetName, query, body)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("create request: %w", err)
	}

	var entity T
	if err := resp.DecodeJSON(&entity); err != nil {
		var zero T
		return zero, fmt.Errorf("decoding created entity: %w", err)
	}

	// Check if there are ANY expands (DefaultExpand or opts.Expand)
	hasExpands := len(a.DefaultExpand) > 0 || (opts != nil && len(opts.Expand) > 0)

	// If no expands, return the entity from POST response
	if !hasExpands {
		a.client.logger.DebugContext(ctx, "Successfully created record.", "entitySetName", a.entitySetName, "id", entity.GetID())
		return entity, nil
	}

	// If expands needed, make GET request to retrieve entity with expands from live DB
	a.client.logger.DebugContext(ctx, "Successfully created record, retrieving with expands.", "entitySetName", a.entitySetName, "id", entity.GetID())
	return a.GetByID(ctx, entity.GetID(), &GetOptions{
		Expand:    opts.Expand,
		Select:    opts.Select,
		UseLiveDB: true,
	})
}

// Delete makes a DELETE request to the endpoint.
// It requires an ID and optionally accepts an ETag for optimistic concurrency.
func (a *APIPage[T]) Delete(ctx context.Context, id uuid.UUID, etag string) error {
	a.client.logger.DebugContext(ctx, "Making DELETE request.", "entitySetName", a.entitySetName, "id", id)

	// Build request options for ETag (if provided)
	var reqOpts []RequestOption
	if etag != "" {
		reqOpts = append(reqOpts, WithETag(etag))
	}

	_, err := a.client.DoRequest(ctx, http.MethodDelete, a.pathWithID(id), nil, nil, reqOpts...)
	if err != nil {
		return fmt.Errorf("DELETE request with id %s: %w", id, err)
	}

	a.client.logger.Debug("Successful DELETE.", "id", id)

	return nil
}

func (a *APIPage[T]) pathWithID(id uuid.UUID) string {
	return fmt.Sprintf("%s(%s)", a.entitySetName, id.String())
}

// mergeExpands merges DefaultExpand with the provided expands.
// Returns DefaultExpand + expands concatenated.
func (a *APIPage[T]) mergeExpands(expands []string) []string {
	if len(a.DefaultExpand) == 0 {
		return expands
	}
	if len(expands) == 0 {
		return a.DefaultExpand
	}
	return slices.Concat(a.DefaultExpand, expands)
}

// mergeFilter merges DefaultFilter with the provided filter using AND logic.
// Returns "(DefaultFilter) and (filter)" or whichever is non-empty.
func (a *APIPage[T]) mergeFilter(filter string) string {
	if a.DefaultFilter == "" {
		return filter
	}
	if filter == "" {
		return a.DefaultFilter
	}
	return fmt.Sprintf("(%s) and (%s)", a.DefaultFilter, filter)
}

