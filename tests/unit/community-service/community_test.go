package community_test

import "testing"

func TestPeoplePageBounds(t *testing.T) {
	page, per := clampPeoplePage(0, 0)
	if page != 1 || per != 12 {
		t.Fatalf("defaults page=%d per=%d", page, per)
	}
	page, per = clampPeoplePage(2, 80)
	if page != 2 || per != 12 {
		t.Fatalf("cap per page=%d per=%d", page, per)
	}
}

func clampPeoplePage(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 12
	}
	return page, perPage
}
