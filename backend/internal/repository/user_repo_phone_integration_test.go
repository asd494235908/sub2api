//go:build integration

package repository

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (s *UserRepoSuite) TestListWithFilters_ReturnsPhoneNumber() {
	user := s.mustCreateUser(&service.User{
		Email:       "phone-list@test.com",
		PhoneNumber: "+8618380640817",
		Username:    "phone-user",
	})

	users, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, service.UserListFilters{})
	s.Require().NoError(err)
	s.Require().NotEmpty(users)

	var matched *service.User
	for i := range users {
		if users[i].ID == user.ID {
			matched = &users[i]
			break
		}
	}
	s.Require().NotNil(matched)
	s.Require().Equal("+8618380640817", matched.PhoneNumber)
}

func (s *UserRepoSuite) TestListWithFilters_HasPhoneOnlyReturnsBoundPhones() {
	matched := s.mustCreateUser(&service.User{
		Email:       "phone-filter-match@test.com",
		PhoneNumber: "+8613800138000",
		Username:    "phone-filter-match",
	})
	s.mustCreateUser(&service.User{
		Email:       "phone-filter-empty@test.com",
		PhoneNumber: "",
		Username:    "phone-filter-empty",
	})
	s.mustCreateUser(&service.User{
		Email:       "phone-filter-space@test.com",
		PhoneNumber: "   ",
		Username:    "phone-filter-space",
	})

	users, page, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, service.UserListFilters{
		HasPhone: true,
	})
	s.Require().NoError(err)
	s.Require().Equal(int64(1), page.Total)
	s.Require().Len(users, 1)
	s.Require().Equal(matched.ID, users[0].ID)
	s.Require().Equal("+8613800138000", users[0].PhoneNumber)
}

func (s *UserRepoSuite) TestListWithFilters_SearchByPhoneNumber() {
	matched := s.mustCreateUser(&service.User{
		Email:       "phone-search-match@test.com",
		PhoneNumber: "+8613800138000",
		Username:    "matched-user",
	})
	s.mustCreateUser(&service.User{
		Email:       "phone-search-miss@test.com",
		PhoneNumber: "+8613900139000",
		Username:    "missed-user",
	})

	users, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, service.UserListFilters{
		Search: "138001380",
	})
	s.Require().NoError(err)
	s.Require().Len(users, 1)
	s.Require().Equal(matched.ID, users[0].ID)
}
