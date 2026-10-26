// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectCreationUsers(t *testing.T) {
	allowedUser := &user.User{ID: 1, Username: "user1"}
	otherUser := &user.User{ID: 2, Username: "user2"}

	setAllowed := func(t *testing.T, value interface{}) {
		original := config.ServiceProjectCreationUsers.GetStringSlice()
		config.ServiceProjectCreationUsers.Set(value)
		t.Cleanup(func() { config.ServiceProjectCreationUsers.Set(original) })
	}

	t.Run("empty list allows everyone", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		setAllowed(t, []string{})

		can, err := (&Project{Title: "p"}).CanCreate(s, otherUser)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("only listed users can create top-level and child projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		setAllowed(t, []string{"USER1, user3"})

		// The inbox created on registration goes through Create, not CanCreate.
		own := &Project{Title: "own"}
		require.NoError(t, own.Create(s, otherUser))

		can, err := (&Project{Title: "p"}).CanCreate(s, allowedUser)
		require.NoError(t, err)
		assert.True(t, can)

		can, err = (&Project{Title: "p"}).CanCreate(s, otherUser)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = (&Project{Title: "child", ParentProjectID: &own.ID}).CanCreate(s, otherUser)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = (&ProjectDuplicate{ProjectID: own.ID}).CanCreate(s, otherUser)
		require.NoError(t, err)
		assert.False(t, can)

		can, err = CanCreateProjects(s, &LinkSharing{ID: 1})
		require.NoError(t, err)
		assert.False(t, can)
	})
}
