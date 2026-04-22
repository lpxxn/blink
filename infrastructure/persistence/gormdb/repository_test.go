package gormdb

import (
	"context"
	"errors"
	"strings"
	"testing"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/lpxxn/blink/internal/testutil"
)

func TestUserRepository_Create_FindByEmail_GetByID(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &UserRepository{DB: db}
	ctx := context.Background()

	u := &domainuser.User{
		SnowflakeID:  9001,
		Email:        "repo-test@example.com",
		Name:         "Repo Test",
		WechatID:     "",
		Phone:        "",
		PasswordHash: "hash",
		PasswordSalt: "",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatal(err)
	}

	byEmail, err := repo.FindByEmail(ctx, "repo-test@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if byEmail.SnowflakeID != 9001 || byEmail.Name != "Repo Test" {
		t.Fatalf("FindByEmail: %+v", byEmail)
	}

	byID, err := repo.GetByID(ctx, 9001)
	if err != nil {
		t.Fatal(err)
	}
	if byID.Email != "repo-test@example.com" {
		t.Fatalf("GetByID: %+v", byID)
	}
}

func TestUserRepository_FindByEmail_GetByID_NotFound(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &UserRepository{DB: db}
	ctx := context.Background()

	_, err := repo.FindByEmail(ctx, "missing@example.com")
	if !errors.Is(err, domainuser.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	_, err = repo.GetByID(ctx, 999999)
	if !errors.Is(err, domainuser.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &UserRepository{DB: db}
	ctx := context.Background()
	base := &domainuser.User{
		SnowflakeID:  9101,
		Email:        "dup@example.com",
		Name:         "A",
		PasswordHash: "h",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}
	if err := repo.Create(ctx, base); err != nil {
		t.Fatal(err)
	}
	dup := &domainuser.User{
		SnowflakeID:  9102,
		Email:        "dup@example.com",
		Name:         "B",
		PasswordHash: "h2",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}
	err := repo.Create(ctx, dup)
	if err == nil {
		t.Fatal("expected error on duplicate email")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Fatalf("expected unique constraint in error, got: %v", err)
	}
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &UserRepository{DB: db}
	ctx := context.Background()
	u := &domainuser.User{
		SnowflakeID:  9201,
		Email:        "login@example.com",
		Name:         "L",
		PasswordHash: "h",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateLastLogin(ctx, 9201, "192.0.2.1", "test-ua"); err != nil {
		t.Fatal(err)
	}
	var m UserModel
	if err := db.WithContext(ctx).Where("snowflake_id = ?", 9201).First(&m).Error; err != nil {
		t.Fatal(err)
	}
	if m.LastLoginAt == nil || m.LastLoginAt.IsZero() {
		t.Fatal("expected last_login_at set")
	}
	if m.LastLoginIP == nil || *m.LastLoginIP != "192.0.2.1" {
		t.Fatalf("last_login_ip: %v", m.LastLoginIP)
	}
	if m.LastLoginDevice == nil || *m.LastLoginDevice != "test-ua" {
		t.Fatalf("last_login_device: %v", m.LastLoginDevice)
	}
}

func TestUserRepository_UpdateLastLogin_EmptyIPDeviceWritesNull(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &UserRepository{DB: db}
	ctx := context.Background()
	u := &domainuser.User{
		SnowflakeID:  9202,
		Email:        "empty-ip@example.com",
		Name:         "E",
		PasswordHash: "h",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateLastLogin(ctx, 9202, "", ""); err != nil {
		t.Fatal(err)
	}
	var m UserModel
	if err := db.WithContext(ctx).Where("snowflake_id = ?", 9202).First(&m).Error; err != nil {
		t.Fatal(err)
	}
	if m.LastLoginAt == nil {
		t.Fatal("expected last_login_at set")
	}
	if m.LastLoginIP != nil {
		t.Fatalf("expected nil last_login_ip, got %v", *m.LastLoginIP)
	}
	if m.LastLoginDevice != nil {
		t.Fatalf("expected nil last_login_device, got %v", *m.LastLoginDevice)
	}
}

func TestUserRepository_SoftDelete_HiddenFromQueries(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &UserRepository{DB: db}
	ctx := context.Background()
	u := &domainuser.User{
		SnowflakeID:  9301,
		Email:        "soft@example.com",
		Name:         "S",
		PasswordHash: "h",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Where("snowflake_id = ?", 9301).Delete(&UserModel{}).Error; err != nil {
		t.Fatal(err)
	}
	_, err := repo.FindByEmail(ctx, "soft@example.com")
	if !errors.Is(err, domainuser.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after soft delete, got %v", err)
	}
	_, err = repo.GetByID(ctx, 9301)
	if !errors.Is(err, domainuser.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after soft delete, got %v", err)
	}
}

func TestOAuthRepository_Create_FindByProviderSubject(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	userRepo := &UserRepository{DB: db}
	oauthRepo := &OAuthRepository{DB: db}
	ctx := context.Background()

	if err := userRepo.Create(ctx, &domainuser.User{
		SnowflakeID:  9401,
		Email:        "oauth-user@example.com",
		Name:         "O",
		PasswordHash: "h",
		Status:       domainuser.StatusActive,
		Role:         "user",
	}); err != nil {
		t.Fatal(err)
	}
	id := &domainoauth.Identity{
		SnowflakeID:     9402,
		Provider:        "test",
		ProviderSubject: "sub-xyz",
		UserID:          9401,
	}
	if err := oauthRepo.Create(ctx, id); err != nil {
		t.Fatal(err)
	}
	got, err := oauthRepo.FindByProviderSubject(ctx, "test", "sub-xyz")
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != 9401 || got.SnowflakeID != 9402 {
		t.Fatalf("identity: %+v", got)
	}
}

func TestOAuthRepository_FindByProviderSubject_NotFound(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	repo := &OAuthRepository{DB: db}
	ctx := context.Background()
	_, err := repo.FindByProviderSubject(ctx, "none", "nope")
	if !errors.Is(err, domainoauth.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestOAuthRepository_DuplicateProviderSubject(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	userRepo := &UserRepository{DB: db}
	oauthRepo := &OAuthRepository{DB: db}
	ctx := context.Background()
	if err := userRepo.Create(ctx, &domainuser.User{
		SnowflakeID: 9501, Email: "u1@example.com", Name: "U1",
		PasswordHash: "h", Status: domainuser.StatusActive, Role: "user",
	}); err != nil {
		t.Fatal(err)
	}
	if err := userRepo.Create(ctx, &domainuser.User{
		SnowflakeID: 9502, Email: "u2@example.com", Name: "U2",
		PasswordHash: "h", Status: domainuser.StatusActive, Role: "user",
	}); err != nil {
		t.Fatal(err)
	}
	first := &domainoauth.Identity{SnowflakeID: 9511, Provider: "p", ProviderSubject: "same", UserID: 9501}
	second := &domainoauth.Identity{SnowflakeID: 9512, Provider: "p", ProviderSubject: "same", UserID: 9502}
	if err := oauthRepo.Create(ctx, first); err != nil {
		t.Fatal(err)
	}
	err := oauthRepo.Create(ctx, second)
	if err == nil {
		t.Fatal("expected error on duplicate (provider, provider_subject)")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Fatalf("expected unique in error: %v", err)
	}
}

func TestOAuthRepository_SoftDelete_Hidden(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	userRepo := &UserRepository{DB: db}
	oauthRepo := &OAuthRepository{DB: db}
	ctx := context.Background()
	if err := userRepo.Create(ctx, &domainuser.User{
		SnowflakeID: 9601, Email: "soft-oauth@example.com", Name: "SO",
		PasswordHash: "h", Status: domainuser.StatusActive, Role: "user",
	}); err != nil {
		t.Fatal(err)
	}
	id := &domainoauth.Identity{
		SnowflakeID: 9602, Provider: "softp", ProviderSubject: "sub-soft", UserID: 9601,
	}
	if err := oauthRepo.Create(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Where("snowflake_id = ?", 9602).Delete(&OAuthIdentityModel{}).Error; err != nil {
		t.Fatal(err)
	}
	_, err := oauthRepo.FindByProviderSubject(ctx, "softp", "sub-soft")
	if !errors.Is(err, domainoauth.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDomainMapping_RoundTrip(t *testing.T) {
	u := &domainuser.User{
		SnowflakeID: 1, Email: "m@x.com", Name: "N", WechatID: "w", Phone: "p",
		PasswordHash: "ph", PasswordSalt: "ps", Status: 1, Role: "user",
	}
	m := domainToUserModel(u)
	out := userModelToDomain(m)
	if out.SnowflakeID != u.SnowflakeID || out.Email != u.Email || out.PasswordHash != u.PasswordHash {
		t.Fatalf("user round-trip: %+v", out)
	}
	oid := &domainoauth.Identity{SnowflakeID: 2, Provider: "a", ProviderSubject: "b", UserID: 1}
	om := identityToModel(oid)
	oid2 := identityModelToDomain(om)
	if oid2.Provider != oid.Provider || oid2.UserID != oid.UserID {
		t.Fatalf("oauth round-trip: %+v", oid2)
	}
}
