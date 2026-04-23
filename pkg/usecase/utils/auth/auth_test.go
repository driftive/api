package auth

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	authmodel "driftive.cloud/api/pkg/model/auth"
	jwtutil "driftive.cloud/api/pkg/usecase/utils/jwt"
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
)

const testJWTSecret = "test-secret-that-is-at-least-32-chars!"

// newTestApp creates a Fiber app with the jwtware middleware wired up
// identically to main.go, and a handler that calls MustGetLoggedUserId.
func newTestApp() *fiber.App {
	app := fiber.New()

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(testJWTSecret)},
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).SendString("unauthorized")
		},
	}))

	app.Get("/protected", func(c fiber.Ctx) error {
		userId, err := MustGetLoggedUserId(c)
		if err != nil {
			return err
		}
		return c.SendString(fmt.Sprintf("%d", *userId))
	})

	return app
}

func TestMustGetLoggedUserId_ValidToken(t *testing.T) {
	app := newTestApp()

	token, err := jwtutil.GenerateJWTToken(authmodel.UserToken{ID: 42, Provider: "GITHUB"}, testJWTSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "42" {
		t.Errorf("expected user id 42, got %s", string(body))
	}
}

func TestMustGetLoggedUserId_NoToken(t *testing.T) {
	app := newTestApp()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMustGetLoggedUserId_InvalidToken(t *testing.T) {
	app := newTestApp()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMustGetLoggedUserId_WrongSecret(t *testing.T) {
	app := newTestApp()

	// Generate token with a different secret
	token, err := jwtutil.GenerateJWTToken(
		authmodel.UserToken{ID: 42, Provider: "GITHUB"},
		"a-completely-different-secret-that-is-long-enough",
	)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}
