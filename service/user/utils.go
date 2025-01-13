package user

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func verifyGoogleToken(accessToken string) (*GoogleUser, error) {
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v3/userinfo?access_token=%s", accessToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("status: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to verify token, status: %d", resp.StatusCode)
	}

	var user GoogleUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

