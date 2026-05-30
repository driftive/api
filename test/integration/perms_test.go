package integration

import (
	"context"
	"sort"
	"testing"

	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/repository"
)

// TestPermsMiddleware_OrgScoping verifies that FindAllUserOrganizationIds —
// the query that backs the perms middleware (pkg/middleware/perms/session.go)
// — only returns org IDs the user actually belongs to. This is the
// org-scoping contract the rest of the authenticated API relies on.
func TestPermsMiddleware_OrgScoping(t *testing.T) {
	truncateAll(t)
	ctx := context.Background()
	pool := withPool(t)

	// Two users.
	insertUser := func(providerID, username string) int64 {
		var id int64
		email := username + "@test"
		err := pool.QueryRow(ctx,
			`INSERT INTO users (provider, provider_id, name, username, email, access_token, refresh_token)
			 VALUES ('GITHUB', $1, $2, $3, $4, 'at', 'rt') RETURNING id`,
			providerID, username, username, email).Scan(&id)
		if err != nil {
			t.Fatalf("insert user %s: %v", username, err)
		}
		return id
	}
	alice := insertUser("100", "alice")
	bob := insertUser("101", "bob")

	// Three orgs.
	insertOrg := func(providerID, name string) int64 {
		var id int64
		err := pool.QueryRow(ctx,
			`INSERT INTO git_organization (provider, provider_id, name) VALUES ('GITHUB', $1, $2) RETURNING id`,
			providerID, name).Scan(&id)
		if err != nil {
			t.Fatalf("insert org %s: %v", name, err)
		}
		return id
	}
	orgA := insertOrg("o-a", "acme")
	orgB := insertOrg("o-b", "beta")
	orgC := insertOrg("o-c", "gamma")

	// alice is in orgA + orgB; bob is in orgC only.
	link := func(userID, orgID int64) {
		if _, err := pool.Exec(ctx,
			`INSERT INTO user_git_organization (user_id, git_organization_id, role) VALUES ($1, $2, 'member')`,
			userID, orgID); err != nil {
			t.Fatalf("link user %d -> org %d: %v", userID, orgID, err)
		}
	}
	link(alice, orgA)
	link(alice, orgB)
	link(bob, orgC)

	repos := repository.NewRepository(testDB, &config.Config{})
	orgRepo := repos.GitOrgRepository()

	aliceOrgs, err := orgRepo.FindAllUserOrganizationIds(ctx, alice)
	if err != nil {
		t.Fatalf("FindAllUserOrganizationIds(alice): %v", err)
	}
	sort.Slice(aliceOrgs, func(i, j int) bool { return aliceOrgs[i] < aliceOrgs[j] })
	want := []int64{orgA, orgB}
	sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
	if len(aliceOrgs) != len(want) {
		t.Fatalf("alice: expected %v, got %v", want, aliceOrgs)
	}
	for i := range want {
		if aliceOrgs[i] != want[i] {
			t.Fatalf("alice: expected %v, got %v", want, aliceOrgs)
		}
	}

	bobOrgs, err := orgRepo.FindAllUserOrganizationIds(ctx, bob)
	if err != nil {
		t.Fatalf("FindAllUserOrganizationIds(bob): %v", err)
	}
	if len(bobOrgs) != 1 || bobOrgs[0] != orgC {
		t.Fatalf("bob: expected [%d], got %v", orgC, bobOrgs)
	}

	// orgC must NOT appear in alice's set — the scoping invariant.
	for _, id := range aliceOrgs {
		if id == orgC {
			t.Errorf("orgC leaked into alice's org IDs: %v", aliceOrgs)
		}
	}
}
