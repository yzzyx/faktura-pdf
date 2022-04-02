package views

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchCurrentPage(t *testing.T) {
	routes := map[string]string{
		"review-list":            "/admin/review",
		"review-schedule-list":   "/admin/review/schedule",
		"review-schedule-create": "/admin/review/schedule/new",
		"review-schedule-view":   "/admin/review/schedule/{id}",
		"role-list":              "/admin/role",
		"adminrole-list":         "/admin/adminrole",
		"actor-list":             "/admin/actor",
		"group-list":             "/admin/group",
		"permission-list":        "/admin/permission",
		"ou-list":                "/admin/ou",
		"costcentre-list":        "/admin/costcentre",
		"delegationrole-list":    "/admin/delegationrole",
		"person-list":            "/admin/person",
		"audit":                  "/admin/audit",
		"base":                   "/",
	}

	type T struct {
		baseURL      string
		path         string
		expectedPage string
	}

	tests := []T{
		{baseURL: "", path: "/admin/review/schedule/new", expectedPage: "review-schedule-create"},
		{baseURL: "", path: "/admin/review/schedule/5", expectedPage: "review-schedule-view"},
		{baseURL: "", path: "/admin/role", expectedPage: "role-list"},
		{baseURL: "", path: "/", expectedPage: "base"},
		{baseURL: "", path: "/admin/review", expectedPage: "review-list"},

		{baseURL: "", path: "/blah", expectedPage: ""},
		{baseURL: "", path: "/admin/review/blah", expectedPage: ""},
		{baseURL: "", path: "/admin/review/schedule/new/blah", expectedPage: ""},

		{baseURL: "/test", path: "/test/admin/review/schedule/new", expectedPage: "review-schedule-create"},
		{baseURL: "/test", path: "/test/admin/review/schedule/5", expectedPage: "review-schedule-view"},
		{baseURL: "/test", path: "/test/admin/role", expectedPage: "role-list"},
		{baseURL: "/test", path: "/test", expectedPage: "base"},
		{baseURL: "/test", path: "/test/admin/review", expectedPage: "review-list"},
	}

	for _, test := range tests {

		cp, err := matchCurrentPage(routes, test.baseURL, test.path)
		require.Nil(t, err)
		require.Equal(t, test.expectedPage, cp, "could not match %s", test.path)
	}

}
