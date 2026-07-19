package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const apiBase = "http://127.0.0.1:8080/api/v1"

func getAuthToken(t *testing.T, username string) string {
	// Register user
	regBody, _ := json.Marshal(map[string]string{
		"username": username,
		"email":    username + "@gitlink.test",
		"password": "password123",
	})
	respReg, err := http.Post(apiBase+"/auth/register", "application/json", bytes.NewBuffer(regBody))
	require.NoError(t, err)
	// Registration might return 201 or 409 if already exists, both are fine for test.
	respReg.Body.Close()

	// Login
	loginBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "password123",
	})
	respLogin, err := http.Post(apiBase+"/auth/login", "application/json", bytes.NewBuffer(loginBody))
	require.NoError(t, err)
	defer respLogin.Body.Close()
	require.Equal(t, http.StatusOK, respLogin.StatusCode)

	var res struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(respLogin.Body).Decode(&res)
	require.NoError(t, err)
	require.NotEmpty(t, res.Token)

	return res.Token
}

func TestE2E_Repository(t *testing.T) {
	username := fmt.Sprintf("user-%d", time.Now().UnixNano())
	token := getAuthToken(t, username)
	repoName := "e2e-test-repo-" + time.Now().Format("150405")

	// 1. Create Repository
	reqBody, _ := json.Marshal(map[string]string{
		"name":        repoName,
		"description": "E2E Integration Test",
	})
	
	req, err := http.NewRequest(http.MethodPost, apiBase+"/repos", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// 2. Duplicate Create
	reqDup, err := http.NewRequest(http.MethodPost, apiBase+"/repos", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	reqDup.Header.Set("Content-Type", "application/json")
	reqDup.Header.Set("Authorization", "Bearer "+token)

	respDup, err := http.DefaultClient.Do(reqDup)
	require.NoError(t, err)
	defer respDup.Body.Close()
	assert.NotEqual(t, http.StatusCreated, respDup.StatusCode, "Duplicate create should fail")

	// 3. Delete Repository
	reqDel, _ := http.NewRequest(http.MethodDelete, apiBase+"/repos/"+repoName, nil)
	reqDel.Header.Set("Authorization", "Bearer "+token)
	respDel, err := http.DefaultClient.Do(reqDel)
	require.NoError(t, err)
	defer respDel.Body.Close()
	assert.Equal(t, http.StatusOK, respDel.StatusCode)
}

func TestE2E_InvalidName(t *testing.T) {
	username := fmt.Sprintf("user-inv-%d", time.Now().UnixNano())
	token := getAuthToken(t, username)

	reqBody, _ := json.Marshal(map[string]string{
		"name": "../invalid-repo",
	})
	
	req, err := http.NewRequest(http.MethodPost, apiBase+"/repos", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusCreated, resp.StatusCode, "Invalid name should fail")
}
