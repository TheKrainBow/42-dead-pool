package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"dead-pool/poolmanager"
)

var clientID = "u-s4t2af-c1e27da8b90b444dd980d733a666bef6eacb4ffaaaf2eff1aae206d2077d1558"
var clientSecret = "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"

type APIProject struct {
	TeamId    int    `json:"id"`
	ProjectId int    `json:"project_id"`
	FinalMark int    `json:"final_mark"`
	Validated bool   `json:"validated?"`
	Users     []User `json:"users"`
}

type User struct {
	ProjectsUserId int `json:"projects_user_id"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

/* Token */
func extractAccessToken(jsonResponse []byte) (string, error) {
	var tokenResp TokenResponse
	if err := json.Unmarshal(jsonResponse, &tokenResp); err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

func getAccessToken() (string, error) {
	url := "https://api.intra.42.fr/oauth/token"
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"scope":         "public projects",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	time.Sleep(time.Millisecond * 501)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return extractAccessToken(body)
}

/* API Projects */
func fetchProject(userID, projectID, accessToken string) (APIProject, error) {
	url := fmt.Sprintf("https://api.intra.42.fr/v2/users/%s/projects/%s/teams?sort=-final_mark", userID, projectID) // Sorted on -final_mark, we only keep the best mark
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return APIProject{TeamId: 0}, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	time.Sleep(time.Millisecond * 501)
	if err != nil {
		return APIProject{TeamId: 0}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return APIProject{TeamId: 0}, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return APIProject{TeamId: 0}, err
	}

	var result []APIProject
	if err := json.Unmarshal(body, &result); err != nil {
		return APIProject{TeamId: 0}, err
	}

	if len(result) == 0 {
		return APIProject{TeamId: 0}, fmt.Errorf("empty response array")
	}
	return result[0], nil
}

func updateProjectTeamMark(teamID int, projectUsersID int, mark int, accessToken string) (err error) {
	urlTeam := fmt.Sprintf("https://api.intra.42.fr/v2/teams/%d", teamID)
	urlProjectSession := fmt.Sprintf("https://api.intra.42.fr/v2/projects_users/%d", projectUsersID)

	// Prepare the payload for the PATCH request
	payload := map[string]interface{}{
		"final_mark": mark,
		"status":     "finished",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	{
		client := &http.Client{}
		req, err := http.NewRequest("PATCH", urlTeam, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create PATCH request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := client.Do(req)
		time.Sleep(time.Millisecond * 501)
		if err != nil {
			return fmt.Errorf("API request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("API request failed with status: %s", resp.Status)
		}
	}
	{
		client := &http.Client{}
		req, err := http.NewRequest("PATCH", urlProjectSession, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create PATCH request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := client.Do(req)
		time.Sleep(time.Millisecond * 501)
		if err != nil {
			return fmt.Errorf("API request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("API request failed with status: %s", resp.Status)
		}
	}
	return nil
}

func calculateParentMark(userID, parentID, accessToken string) (mark int, err error) {
	childrenIDs := poolmanager.GetProjectChildrenIDs(parentID)
	mark = 0
	if len(childrenIDs) == 0 {
		return mark, nil
	}
	for _, id := range childrenIDs {
		child, err := fetchProject(userID, id, accessToken)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch child module %s: %s", id, err.Error())
		}
		if !child.Validated {
			return 0, fmt.Errorf("piscine module is not validated: %d", child.ProjectId)
		}
		mark += child.FinalMark
	}
	mark /= len(childrenIDs)
	return mark, nil
}

func UpdatePoolParentProject(userID string, moduleID string) (err error) {
	// Load the configuration from the file
	err = poolmanager.LoadConfig("pool-list.json")
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	parentID := poolmanager.GetProjectParentID(moduleID)
	if parentID == "" {
		return fmt.Errorf("invalid ProjectID provided (%s) is not a known pool module", moduleID)
	}

	accessToken, err := getAccessToken()
	if err != nil {
		return fmt.Errorf("error getting access token: %w", err)
	}

	mark, err := calculateParentMark(userID, parentID, accessToken)
	if err != nil {
		return fmt.Errorf("error calculating final mark: %w", err)
	}

	parent, err := fetchProject(userID, parentID, accessToken)
	if err != nil {
		return fmt.Errorf("error fetching data: %w", err)
	}

	if parent.FinalMark > mark {
		return fmt.Errorf("current mark for parent (%d) is higher than new mark (%d)", parent.FinalMark, mark)
	}
	err = updateProjectTeamMark(parent.TeamId, parent.Users[0].ProjectsUserId, mark, accessToken)
	if err != nil {
		return fmt.Errorf("error updating data: %w", err)
	}
	fmt.Printf("Succesfully updated team %d's mark to %d.\n", parent.TeamId, mark)
	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: program <userID> <projectID>")
		return
	}
	if clientID == "" || clientSecret == "" {
		fmt.Println("You must provide a clientID and a clientSecret.")
		return
	}

	err := UpdatePoolParentProject(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println("Update aborted: %w", err.Error())
	}
}
