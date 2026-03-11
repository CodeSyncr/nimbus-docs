package controllers

import (
	"github.com/CodeSyncr/nimbus/http"
	"github.com/CodeSyncr/nimbus/plugins/unpoly"
)

// HelloWorld controller implements router.ResourceController.
type HelloWorld struct{}

func (c *HelloWorld) Index(ctx *http.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"message": "Hello World index"})
}

func (c *HelloWorld) Create(ctx *http.Context) error {
	ctx.String(http.StatusOK, "Show create form")
	return nil
}

func (c *HelloWorld) Store(ctx *http.Context) error {
	if unpoly.IsValidating(ctx) {
		fields := unpoly.ValidateNames(ctx)
		errors := make(map[string]string)
		for _, f := range fields {
			if f == "name" {
				errors[f] = ""
			}
		}
		return ctx.JSON(http.StatusOK, errors)
	}

	unpoly.EmitEvent(ctx, "hello:created", map[string]any{"message": "Created"})
	unpoly.ExpireCache(ctx, "/hello*")
	return ctx.JSON(http.StatusCreated, map[string]string{"message": "Created"})
}

func (c *HelloWorld) Show(ctx *http.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(http.StatusOK, map[string]string{"id": id, "message": "Hello World show"})
}

func (c *HelloWorld) Edit(ctx *http.Context) error {
	id := ctx.Param("id")
	ctx.String(http.StatusOK, "Edit form for "+id)
	return nil
}

func (c *HelloWorld) Update(ctx *http.Context) error {
	id := ctx.Param("id")

	unpoly.EmitEvent(ctx, "hello:updated", map[string]any{"id": id})
	unpoly.ExpireCache(ctx, "/hello*")
	return ctx.JSON(http.StatusOK, map[string]string{"id": id, "message": "Updated"})
}

func (c *HelloWorld) Destroy(ctx *http.Context) error {
	id := ctx.Param("id")

	unpoly.EmitEvent(ctx, "hello:destroyed", map[string]any{"id": id})
	unpoly.EvictCache(ctx, "/hello*")

	if unpoly.IsUnpoly(ctx) {
		unpoly.AcceptLayer(ctx, map[string]string{"deleted": id})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"id": id, "message": "Deleted"})
}
