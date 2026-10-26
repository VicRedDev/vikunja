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
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// projectCreationUsers returns the lowercased usernames from service.projectcreationusers.
// Entries may also be comma-separated, so the env var works as "user1,user2".
func projectCreationUsers() map[string]struct{} {
	allowed := map[string]struct{}{}
	for _, entry := range config.ServiceProjectCreationUsers.GetStringSlice() {
		for _, name := range strings.Split(entry, ",") {
			name = strings.ToLower(strings.TrimSpace(name))
			if name != "" {
				allowed[name] = struct{}{}
			}
		}
	}
	return allowed
}

// CanCreateProjects reports whether a may create projects at all. With
// service.projectcreationusers empty everyone may; otherwise only the listed
// users (and bots owned by them) can.
func CanCreateProjects(s *xorm.Session, a web.Auth) (bool, error) {
	allowed := projectCreationUsers()
	if len(allowed) == 0 {
		return true, nil
	}
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return false, err
	}
	u, err := user.GetUserByID(s, doer.ID)
	if err != nil {
		return false, err
	}
	if u.IsBot() {
		u, err = user.GetUserByID(s, u.BotOwnerID)
		if err != nil {
			return false, err
		}
	}

	_, ok := allowed[strings.ToLower(u.Username)]
	return ok, nil
}
