package domain

import "testing"

func TestNewPaginationOptions(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		perPage     int
		sortBy      string
		sortDir     string
		search      string
		wantPage    int
		wantPerPage int
		wantSortBy  string
		wantSortDir string
	}{
		{
			name:        "default values",
			page:        0,
			perPage:     0,
			sortBy:      "",
			sortDir:     "",
			search:      "",
			wantPage:    1,
			wantPerPage: 10,
			wantSortBy:  "created_at",
			wantSortDir: "desc",
		},
		{
			name:        "custom values",
			page:        5,
			perPage:     20,
			sortBy:      "name",
			sortDir:     "asc",
			search:      "test",
			wantPage:    5,
			wantPerPage: 20,
			wantSortBy:  "name",
			wantSortDir: "asc",
		},
		{
			name:        "max per page",
			page:        1,
			perPage:     500,
			sortBy:      "",
			sortDir:     "",
			search:      "",
			wantPage:    1,
			wantPerPage: 100,
			wantSortBy:  "created_at",
			wantSortDir: "desc",
		},
		{
			name:        "negative page",
			page:        -1,
			perPage:     10,
			sortBy:      "",
			sortDir:     "",
			search:      "",
			wantPage:    1,
			wantPerPage: 10,
			wantSortBy:  "created_at",
			wantSortDir: "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewPaginationOptions(tt.page, tt.perPage, tt.sortBy, tt.sortDir, tt.search)

			if opts.Page != tt.wantPage {
				t.Errorf("Page = %v, want %v", opts.Page, tt.wantPage)
			}
			if opts.PerPage != tt.wantPerPage {
				t.Errorf("PerPage = %v, want %v", opts.PerPage, tt.wantPerPage)
			}
			if opts.SortBy != tt.wantSortBy {
				t.Errorf("SortBy = %v, want %v", opts.SortBy, tt.wantSortBy)
			}
			if opts.SortDir != tt.wantSortDir {
				t.Errorf("SortDir = %v, want %v", opts.SortDir, tt.wantSortDir)
			}
			if opts.Search != tt.search {
				t.Errorf("Search = %v, want %v", opts.Search, tt.search)
			}
		})
	}
}

func TestPaginationOptions_Offset(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		perPage    int
		wantOffset int64
	}{
		{"page 1", 1, 10, 0},
		{"page 2", 2, 10, 10},
		{"page 3 with 20 per page", 3, 20, 40},
		{"page 5 with 25 per page", 5, 25, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := PaginationOptions{Page: tt.page, PerPage: tt.perPage}
			if got := opts.Offset(); got != tt.wantOffset {
				t.Errorf("Offset() = %v, want %v", got, tt.wantOffset)
			}
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		perPage  int
		wantPages int
	}{
		{"zero total", 0, 10, 0},
		{"exact pages", 20, 10, 2},
		{"remainder", 25, 10, 3},
		{"single page", 5, 10, 1},
		{"many pages", 101, 10, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateTotalPages(tt.total, tt.perPage); got != tt.wantPages {
				t.Errorf("CalculateTotalPages() = %v, want %v", got, tt.wantPages)
			}
		})
	}
}
