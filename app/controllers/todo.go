package controllers

import (
	"strconv"
	"strings"

	"github.com/CodeSyncr/nimbus"
	"github.com/CodeSyncr/nimbus/database"
	"github.com/CodeSyncr/nimbus/http"
	"github.com/CodeSyncr/nimbus/plugins/unpoly"

	"nimbus-starter/app/models"
	"nimbus-starter/app/validators"
)

// Todo controller implements router.ResourceController.
// It depends on the higher-level *nimbus.DB abstraction, not gorm.
type Todo struct {
	DB *nimbus.DB
}

func (todo *Todo) Index(ctx *http.Context) error {
	var items []models.Todo
	todo.DB.Find(&items)
	if unpoly.IsUnpoly(ctx) {
		unpoly.SetTitle(ctx, "Todo · Nimbus Demos")
	}
	return ctx.View("apps/todo/index", map[string]any{
		"title": "Todo",
		"items": items,
		"empty": len(items) == 0,
	})
}

func (todo *Todo) Create(ctx *http.Context) error {
	if unpoly.IsUnpoly(ctx) {
		unpoly.SetTitle(ctx, "New Todo · Nimbus Demos")
	}
	return ctx.View("apps/todo/form", map[string]any{
		"title": "New Todo",
		"item":  nil,
	})
}

func (todo *Todo) Store(ctx *http.Context) error {
	_ = ctx.Request.ParseForm()
	v := &validators.Todo{Title: strings.TrimSpace(ctx.Request.FormValue("title"))}
	if err := v.Validate(); err != nil {
		return ctx.View("apps/todo/form", map[string]any{
			"title": "New Todo",
			"item":  nil,
			"error": "Title is required (1–255 chars)",
		})
	}
	item := &models.Todo{Title: v.Title, Done: false}
	todo.DB.Create(item)
	ctx.Redirect(http.StatusFound, "/demos/todo")
	return nil
}

func (todo *Todo) Show(ctx *http.Context) error {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	var item models.Todo
	if database.Get().First(&item, id).Error != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	if unpoly.IsUnpoly(ctx) {
		unpoly.SetTitle(ctx, item.Title+" · Nimbus Demos")
	}
	return ctx.View("apps/todo/show", map[string]any{
		"title": "Todo",
		"item":  item,
	})
}

func (todo *Todo) Edit(ctx *http.Context) error {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	var item models.Todo
	if database.Get().First(&item, id).Error != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	if unpoly.IsUnpoly(ctx) {
		unpoly.SetTitle(ctx, "Edit "+item.Title+" · Nimbus Demos")
	}
	return ctx.View("apps/todo/form", map[string]any{
		"title": "Edit Todo",
		"item":  item,
	})
}

func (todo *Todo) Update(ctx *http.Context) error {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	var item models.Todo
	if database.Get().First(&item, id).Error != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	_ = ctx.Request.ParseForm()
	v := &validators.Todo{Title: strings.TrimSpace(ctx.Request.FormValue("title"))}
	if err := v.Validate(); err != nil {
		return ctx.View("apps/todo/form", map[string]any{
			"title": "Edit Todo",
			"item":  item,
			"error": "Title is required (1–255 chars)",
		})
	}
	done := ctx.Request.FormValue("done") == "on"
	database.Get().Model(&item).Updates(map[string]any{"title": v.Title, "done": done})
	ctx.Redirect(http.StatusFound, "/demos/todo")
	return nil
}

func (todo *Todo) Destroy(ctx *http.Context) error {
	id := ctx.Param("id")
	if database.Get().Delete(&models.Todo{}, id).RowsAffected == 0 {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	if unpoly.IsUnpoly(ctx) {
		unpoly.AcceptLayer(ctx, map[string]string{"deleted": id})
	}
	ctx.Redirect(http.StatusFound, "/demos/todo")
	return nil
}
