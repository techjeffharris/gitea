// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"testing"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/util"

	"github.com/stretchr/testify/assert"
)

func TestGetEmailAddresses(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	emails, _ := GetEmailAddresses(int64(1))
	if assert.Len(t, emails, 3) {
		assert.True(t, emails[0].IsPrimary)
		assert.True(t, emails[2].IsActivated)
		assert.False(t, emails[2].IsPrimary)
	}

	emails, _ = GetEmailAddresses(int64(2))
	if assert.Len(t, emails, 2) {
		assert.True(t, emails[0].IsPrimary)
		assert.True(t, emails[0].IsActivated)
	}
}

func TestIsEmailUsed(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	isExist, _ := IsEmailUsed("")
	assert.True(t, isExist)
	isExist, _ = IsEmailUsed("user11@example.com")
	assert.True(t, isExist)
	isExist, _ = IsEmailUsed("user1234567890@example.com")
	assert.False(t, isExist)
}

func TestAddEmailAddress(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	assert.NoError(t, AddEmailAddress(&EmailAddress{
		Email:       "user1234567890@example.com",
		LowerEmail:  "user1234567890@example.com",
		IsPrimary:   true,
		IsActivated: true,
	}))

	// ErrEmailAlreadyUsed
	err := AddEmailAddress(&EmailAddress{
		Email:      "user1234567890@example.com",
		LowerEmail: "user1234567890@example.com",
	})
	assert.Error(t, err)
	assert.True(t, IsErrEmailAlreadyUsed(err))
}

func TestAddEmailAddresses(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	// insert multiple email address
	emails := make([]*EmailAddress, 2)
	emails[0] = &EmailAddress{
		Email:       "user1234@example.com",
		LowerEmail:  "user1234@example.com",
		IsActivated: true,
	}
	emails[1] = &EmailAddress{
		Email:       "user5678@example.com",
		LowerEmail:  "user5678@example.com",
		IsActivated: true,
	}
	assert.NoError(t, AddEmailAddresses(emails))

	// ErrEmailAlreadyUsed
	err := AddEmailAddresses(emails)
	assert.Error(t, err)
	assert.True(t, IsErrEmailAlreadyUsed(err))
}

func TestDeleteEmailAddress(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	assert.NoError(t, DeleteEmailAddress(&EmailAddress{
		UID:        int64(1),
		ID:         int64(33),
		Email:      "user1-2@example.com",
		LowerEmail: "user1-2@example.com",
	}))

	assert.NoError(t, DeleteEmailAddress(&EmailAddress{
		UID:        int64(1),
		Email:      "user1-3@example.com",
		LowerEmail: "user1-3@example.com",
	}))

	// Email address does not exist
	err := DeleteEmailAddress(&EmailAddress{
		UID:        int64(1),
		Email:      "user1234567890@example.com",
		LowerEmail: "user1234567890@example.com",
	})
	assert.Error(t, err)
}

func TestDeleteEmailAddresses(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	// delete multiple email address
	emails := make([]*EmailAddress, 2)
	emails[0] = &EmailAddress{
		UID:        int64(2),
		ID:         int64(3),
		Email:      "user2@example.com",
		LowerEmail: "user2@example.com",
	}
	emails[1] = &EmailAddress{
		UID:        int64(2),
		Email:      "user2-2@example.com",
		LowerEmail: "user2-2@example.com",
	}
	assert.NoError(t, DeleteEmailAddresses(emails))

	// ErrEmailAlreadyUsed
	err := DeleteEmailAddresses(emails)
	assert.Error(t, err)
}

func TestMakeEmailPrimary(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	email := &EmailAddress{
		Email: "user567890@example.com",
	}
	err := MakeEmailPrimary(email)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrEmailAddressNotExist{email.Email}.Error())

	email = &EmailAddress{
		Email: "user11@example.com",
	}
	err = MakeEmailPrimary(email)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrEmailNotActivated.Error())

	email = &EmailAddress{
		Email: "user9999999@example.com",
	}
	err = MakeEmailPrimary(email)
	assert.Error(t, err)
	assert.True(t, IsErrUserNotExist(err))

	email = &EmailAddress{
		Email: "user101@example.com",
	}
	err = MakeEmailPrimary(email)
	assert.NoError(t, err)

	user, _ := GetUserByID(int64(10))
	assert.Equal(t, "user101@example.com", user.Email)
}

func TestActivate(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	email := &EmailAddress{
		ID:    int64(1),
		UID:   int64(1),
		Email: "user11@example.com",
	}
	assert.NoError(t, email.Activate())

	emails, _ := GetEmailAddresses(int64(1))
	assert.Len(t, emails, 3)
	assert.True(t, emails[0].IsActivated)
	assert.True(t, emails[0].IsPrimary)
	assert.False(t, emails[1].IsPrimary)
	assert.True(t, emails[2].IsActivated)
	assert.False(t, emails[2].IsPrimary)
}

func TestListEmails(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())

	// Must find all users and their emails
	opts := &SearchEmailOptions{
		ListOptions: db.ListOptions{
			PageSize: 10000,
		},
	}
	emails, count, err := SearchEmails(opts)
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), count)
	assert.True(t, count > 5)

	contains := func(match func(s *SearchEmailResult) bool) bool {
		for _, v := range emails {
			if match(v) {
				return true
			}
		}
		return false
	}

	assert.True(t, contains(func(s *SearchEmailResult) bool { return s.UID == 18 }))
	// 'user3' is an organization
	assert.False(t, contains(func(s *SearchEmailResult) bool { return s.UID == 3 }))

	// Must find no records
	opts = &SearchEmailOptions{Keyword: "NOTFOUND"}
	emails, count, err = SearchEmails(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Must find users 'user2', 'user28', etc.
	opts = &SearchEmailOptions{Keyword: "user2"}
	emails, count, err = SearchEmails(opts)
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), count)
	assert.True(t, contains(func(s *SearchEmailResult) bool { return s.UID == 2 }))
	assert.True(t, contains(func(s *SearchEmailResult) bool { return s.UID == 27 }))

	// Must find only primary addresses (i.e. from the `user` table)
	opts = &SearchEmailOptions{IsPrimary: util.OptionalBoolTrue}
	emails, count, err = SearchEmails(opts)
	assert.NoError(t, err)
	assert.True(t, contains(func(s *SearchEmailResult) bool { return s.IsPrimary }))
	assert.False(t, contains(func(s *SearchEmailResult) bool { return !s.IsPrimary }))

	// Must find only inactive addresses (i.e. not validated)
	opts = &SearchEmailOptions{IsActivated: util.OptionalBoolFalse}
	emails, count, err = SearchEmails(opts)
	assert.NoError(t, err)
	assert.True(t, contains(func(s *SearchEmailResult) bool { return !s.IsActivated }))
	assert.False(t, contains(func(s *SearchEmailResult) bool { return s.IsActivated }))

	// Must find more than one page, but retrieve only one
	opts = &SearchEmailOptions{
		ListOptions: db.ListOptions{
			PageSize: 5,
			Page:     1,
		},
	}
	emails, count, err = SearchEmails(opts)
	assert.NoError(t, err)
	assert.Len(t, emails, 5)
	assert.Greater(t, count, int64(len(emails)))
}
