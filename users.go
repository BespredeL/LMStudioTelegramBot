package main

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

type BotUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Allowed  bool   `json:"allowed"`
}

var (
	users         = make(map[int64]*BotUser)
	usersFileName = "users.json"
	usersMutex    sync.Mutex
)

// File download from a file
func loadUsers() error {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	data, err := os.ReadFile(usersFileName)
	if err != nil {
		if os.IsNotExist(err) {
			users = make(map[int64]*BotUser)
			return nil
		}
		return err
	}

	var list []*BotUser
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	users = make(map[int64]*BotUser)
	for _, u := range list {
		users[u.ID] = u
	}

	return nil
}

// Making users to file
func saveUsers() error {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	var list []*BotUser
	for _, u := range users {
		list = append(list, u)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(usersFileName, data, 0644)
}

// Adding or updating user
func addOrUpdateUser(userID int64, username string) *BotUser {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	if u, exists := users[userID]; exists {
		u.Username = username
		return u
	}

	newUser := &BotUser{
		ID:       userID,
		Username: username,
		Allowed:  false, // By default, access is prohibited
	}

	users[userID] = newUser
	return newUser
}

// Function for obtaining a sorted user list (for GUI)
func getSortedUsers() []*BotUser {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	var userSlice []*BotUser
	for _, u := range users {
		userSlice = append(userSlice, u)
	}

	sort.Slice(userSlice, func(i, j int) bool {
		return userSlice[i].ID < userSlice[j].ID
	})

	return userSlice
}
