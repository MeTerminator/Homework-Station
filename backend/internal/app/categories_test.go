package app

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"

	"homework-station/internal/config"
	"homework-station/internal/model"
)

func TestNestedCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "test-session-secret-32-characters-long")
	application, err := New(config.Load())
	if err != nil {
		t.Fatal(err)
	}
	db := application.DB
	defer func() {
		application.Close()
	}()
	admin := &testClient{t: t, r: application.Router}
	admin.request("POST", "/api/login", gin.H{"username": "admin", "password": "123456"}, 200)
	create := func(name string, parent *uint) model. // Reparenting is unrestricted by depth, including a whole existing subtree.
		// would form a cycle across the moved subtree
		Category {
		return decode[model.Category](t, admin.request("POST", "/api/admin/categories", gin.H{"name": name, "parent_id": parent}, 200))
	}
	root := create("一级", nil)
	leaf := root
	for depth := 2; depth <= 6; depth++ {
		leaf = create(fmt.Sprintf("第%d级", depth), &leaf.ID)
	}
	other := create("另一一级", nil)
	otherChild := create("另一二级", &other.ID)
	move := func(c model.Category, parent uint, want int) {
		admin.request("PUT", fmt.Sprintf("/api/admin/categories/%d", c.ID), gin.H{"name": c.Name, "parent_id": parent}, want)
	}
	move(root, otherChild.ID, 200)
	move(other, leaf.ID, 400)
	move(root, root.ID, 400)
	var unchanged model.Category
	db.First(&unchanged, other.ID)
	if unchanged.ParentID != nil {
		t.Fatal("rejected cyclic move changed the category")
	}
	result := decode[[]model.Category](t, admin.request("GET", "/api/categories", nil, 200))
	if len(result) != 8 {
		t.Fatalf("category records lost or extra records persisted: %d", len(result))
	}
}
