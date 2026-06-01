// client/main.go
package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/uvamsi76/grpc_kiddy_proj/gen/user"
)

func main() {
	// ── 1. Connect ───────────────────────────────────────
	conn := newConnection("localhost:50051")
	defer conn.Close() // always close when done

	// ── 2. Create client ─────────────────────────────────
	userClient := NewUserClient(pb.NewUserServiceClient(conn))

	ctx := context.Background()

	// ── 3. Create users ──────────────────────────────────
	fmt.Println("\n=== Creating Users ===")

	arjun, err := userClient.CreateUser(ctx,
		"Arjun Kumar", "arjun@example.com", 28, pb.UserRole_USER_ROLE_ADMIN,
	)
	if err != nil {
		log.Fatalf("CreateUser failed: %v", err)
	}
	fmt.Printf("Created: %s (id: %s)\n", arjun.Name, arjun.Id)

	priya, err := userClient.CreateUser(ctx,
		"Priya Sharma", "priya@example.com", 25, pb.UserRole_USER_ROLE_EDITOR,
	)
	if err != nil {
		log.Fatalf("CreateUser failed: %v", err)
	}
	fmt.Printf("Created: %s (id: %s)\n", priya.Name, priya.Id)

	// ── 4. Try duplicate email ───────────────────────────
	fmt.Println("\n=== Testing Duplicate Email ===")
	_, err = userClient.CreateUser(ctx,
		"Arjun Duplicate", "arjun@example.com", 28, pb.UserRole_USER_ROLE_VIEWER,
	)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		// [CreateUser] already exists: a user with email arjun@example.com already exists
	}

	// ── 5. Get user ──────────────────────────────────────
	fmt.Println("\n=== Getting User ===")
	user, err := userClient.GetUser(ctx, arjun.Id)
	if err != nil {
		log.Fatalf("GetUser failed: %v", err)
	}
	fmt.Printf("Got: %s, email: %s, active: %v\n", user.Name, user.Email, user.Active)

	// ── 6. Try get non-existent user ─────────────────────
	fmt.Println("\n=== Testing Not Found ===")
	_, err = userClient.GetUser(ctx, "fake_id_999")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		// [GetUser] not found: user fake_id_999 not found
	}

	// ── 7. Update user (partial) ─────────────────────────
	fmt.Println("\n=== Updating User ===")
	name := "Arjun Kumar (Updated)"
	updated, err := userClient.UpdateUser(ctx, arjun.Id, UpdateUserParams{
		Name: &name, // update name
		// Age and Active are nil — not touched
	})
	if err != nil {
		log.Fatalf("UpdateUser failed: %v", err)
	}
	fmt.Printf("Updated name: %s\n", updated.Name)

	// Deactivate user
	active := false
	updated, err = userClient.UpdateUser(ctx, priya.Id, UpdateUserParams{
		Active: &active,
	})
	if err != nil {
		log.Fatalf("UpdateUser failed: %v", err)
	}
	fmt.Printf("Priya active: %v\n", updated.Active)

	// ── 8. List users ────────────────────────────────────
	fmt.Println("\n=== Listing Users ===")
	users, total, err := userClient.ListUsers(ctx, 1, 10)
	if err != nil {
		log.Fatalf("ListUsers failed: %v", err)
	}
	fmt.Printf("Total users: %d\n", total)
	for _, u := range users {
		fmt.Printf("  - %s (%s) active=%v\n", u.Name, u.Email, u.Active)
	}

	// ── 9. Delete user ───────────────────────────────────
	fmt.Println("\n=== Deleting User ===")
	ok, err := userClient.DeleteUser(ctx, priya.Id)
	if err != nil {
		log.Fatalf("DeleteUser failed: %v", err)
	}
	fmt.Printf("Deleted Priya: %v\n", ok)

	// Confirm deletion
	_, err = userClient.GetUser(ctx, priya.Id)
	if err != nil {
		fmt.Printf("Confirmed deleted: %v\n", err)
	}
}
