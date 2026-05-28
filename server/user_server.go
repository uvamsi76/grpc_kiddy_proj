// server/user_server.go
package main

import (
	"context"
	"log"

	pb "github.com/uvamsi76/grpc_kiddy_proj/gen/user"
	"github.com/uvamsi76/grpc_kiddy_proj/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServer implements the generated UserServiceServer interface
type UserServer struct {
	pb.UnimplementedUserServiceServer // future-proofs against new RPCs
	store                             *store.UserStore
}

func NewUserServer() *UserServer {
	return &UserServer{
		store: store.NewUserStore(),
	}
}

func (s *UserServer) CreateUser(
	ctx context.Context,
	req *pb.CreateUserRequest,
) (*pb.CreateUserResponse, error) {

	// ── Validation ──────────────────────────────────────
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Age < 0 || req.Age > 150 {
		return nil, status.Errorf(codes.InvalidArgument, "age %d is invalid", req.Age)
	}

	// ── Uniqueness check ────────────────────────────────
	if _, exists := s.store.GetByEmail(req.Email); exists {
		return nil, status.Errorf(codes.AlreadyExists,
			"a user with email %s already exists", req.Email)
	}

	// ── Create ──────────────────────────────────────────
	user := s.store.Create(&pb.User{
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
		Role:  req.Role,
	})

	log.Printf("[CreateUser] created user id=%s email=%s", user.Id, user.Email)

	return &pb.CreateUserResponse{User: user}, nil
}

func (s *UserServer) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.GetUserResponse, error) {

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	user, ok := s.store.GetByID(req.Id)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "user %s not found", req.Id)
	}

	return &pb.GetUserResponse{User: user}, nil
}

func (s *UserServer) UpdateUser(
	ctx context.Context,
	req *pb.UpdateUserRequest,
) (*pb.UpdateUserResponse, error) {

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Use optional fields — only update what was explicitly provided
	// req.Name is *string — nil means "don't update this field"
	user, ok := s.store.Update(req.Id, func(u *pb.User) {
		if req.Name != nil {
			u.Name = req.GetName() // GetName() safely unwraps *string
		}
		if req.Age != nil {
			u.Age = req.GetAge()
		}
		if req.Active != nil {
			u.Active = req.GetActive()
		}
	})
	if !ok {
		return nil, status.Errorf(codes.NotFound, "user %s not found", req.Id)
	}

	log.Printf("[UpdateUser] updated user id=%s", user.Id)
	return &pb.UpdateUserResponse{User: user}, nil
}

func (s *UserServer) DeleteUser(
	ctx context.Context,
	req *pb.DeleteUserRequest,
) (*pb.DeleteUserResponse, error) {

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	ok := s.store.Delete(req.Id)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "user %s not found", req.Id)
	}

	log.Printf("[DeleteUser] deleted user id=%s", req.Id)
	return &pb.DeleteUserResponse{Success: true}, nil
}

func (s *UserServer) ListUsers(
	ctx context.Context,
	req *pb.ListUsersRequest,
) (*pb.ListUsersResponse, error) {

	// Apply sensible defaults
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		return nil, status.Error(codes.InvalidArgument, "page_size cannot exceed 100")
	}

	users, total := s.store.List(page, pageSize)

	return &pb.ListUsersResponse{
		Users: users,
		Total: total,
	}, nil
}
