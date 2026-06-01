// client/user_client.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/uvamsi76/grpc_kiddy_proj/gen/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserClient wraps the generated gRPC client with a clean API
type UserClient struct {
	client  pb.UserServiceClient
	timeout time.Duration // default timeout for all calls
}

func NewUserClient(client pb.UserServiceClient) *UserClient {
	return &UserClient{
		client:  client,
		timeout: 5 * time.Second, // sane default
	}
}

// withTimeout creates a context with the default timeout
// but respects an already-cancelled parent context
func (c *UserClient) withTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, c.timeout)
}

func (c *UserClient) CreateUser(
	ctx context.Context,
	name, email string,
	age int32,
	role pb.UserRole,
) (*pb.User, error) {

	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.client.CreateUser(ctx, &pb.CreateUserRequest{
		Name:  name,
		Email: email,
		Age:   age,
		Role:  role,
	})
	if err != nil {
		return nil, handleError("CreateUser", err)
	}

	return resp.User, nil
}
func (c *UserClient) GetUser(ctx context.Context, id string) (*pb.User, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{Id: id})
	if err != nil {
		return nil, handleError("GetUser", err)
	}

	return resp.User, nil
}

// UpdateUserParams uses pointers so callers only set what they want to change
// nil means "don't update this field"
type UpdateUserParams struct {
	Name   *string
	Age    *int32
	Active *bool
}

func (c *UserClient) UpdateUser(
	ctx context.Context,
	id string,
	params UpdateUserParams,
) (*pb.User, error) {

	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	req := &pb.UpdateUserRequest{Id: id}

	// Only set fields the caller wants to change
	if params.Name != nil {
		req.Name = params.Name
	}
	if params.Age != nil {
		req.Age = params.Age
	}
	if params.Active != nil {
		req.Active = params.Active
	}

	resp, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		return nil, handleError("UpdateUser", err)
	}

	return resp.User, nil
}

func (c *UserClient) DeleteUser(ctx context.Context, id string) (bool, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
	if err != nil {
		return false, handleError("DeleteUser", err)
	}

	return resp.Success, nil
}

func (c *UserClient) ListUsers(
	ctx context.Context,
	page, pageSize int32,
) ([]*pb.User, int32, error) {

	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.client.ListUsers(ctx, &pb.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, 0, handleError("ListUsers", err)
	}

	return resp.Users, resp.Total, nil
}

// handleError translates gRPC errors into meaningful messages
// and decides whether to log, wrap, or surface them
func handleError(method string, err error) error {
	if err == nil {
		return nil
	}

	// Extract the gRPC status from the error
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error (network issue, context cancelled, etc.)
		return fmt.Errorf("[%s] non-gRPC error: %w", method, err)
	}

	// Handle each code differently
	switch st.Code() {

	case codes.NotFound:
		// Don't log — caller should handle this
		return fmt.Errorf("[%s] not found: %s", method, st.Message())

	case codes.InvalidArgument:
		// Caller bug — log it
		log.Printf("[%s] invalid argument: %s", method, st.Message())
		return fmt.Errorf("[%s] invalid argument: %s", method, st.Message())

	case codes.AlreadyExists:
		return fmt.Errorf("[%s] already exists: %s", method, st.Message())

	case codes.DeadlineExceeded:
		log.Printf("[%s] deadline exceeded — server too slow or network issue", method)
		return fmt.Errorf("[%s] request timed out", method)

	case codes.Canceled:
		// Caller cancelled — not an error to log
		return fmt.Errorf("[%s] request was cancelled", method)

	case codes.Unavailable:
		// Server down — worth logging and potentially retrying
		log.Printf("[%s] server unavailable: %s", method, st.Message())
		return fmt.Errorf("[%s] server unavailable, try again later", method)

	case codes.Unauthenticated:
		return fmt.Errorf("[%s] not authenticated: %s", method, st.Message())

	case codes.PermissionDenied:
		return fmt.Errorf("[%s] permission denied: %s", method, st.Message())

	case codes.Internal:
		// Server bug — always log
		log.Printf("[%s] internal server error: %s", method, st.Message())
		return fmt.Errorf("[%s] internal server error", method)

	default:
		log.Printf("[%s] unexpected gRPC error code=%s msg=%s", method, st.Code(), st.Message())
		return fmt.Errorf("[%s] unexpected error: %s", method, st.Message())
	}
}

// ── Three levels of timeout control ──────────────────────

// Level 1: Default timeout on the client wrapper (already done in 5.4)
// client := NewUserClient(pb.NewUserServiceClient(conn))
// // Every call gets 5s by default

// // Level 2: Override for a specific call
// func (c *UserClient) GetUserFast(ctx context.Context, id string) (*pb.User, error) {
//     // This specific call only gets 500ms
//     ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
//     defer cancel()

//     resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{Id: id})
//     // ...
// }

// // Level 3: Deadline from the caller (propagated through the call chain)
// func HandleHTTPRequest(w http.ResponseWriter, r *http.Request) {
//     // Honour the HTTP request's deadline end-to-end
//     ctx := r.Context() // already has a deadline from the HTTP server

//     user, err := userClient.GetUser(ctx, r.URL.Query().Get("id"))
//     // The gRPC call respects the HTTP request's deadline automatically
// }
