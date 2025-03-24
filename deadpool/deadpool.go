package deadpool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"dead-pool/poolmanager"
)

var clientID = ""
var clientSecret = ""
var logFile *os.File
var level log.Level = log.InfoLevel

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

func InitLogs(logPath string) (err error) {
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(logFile)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(level)
	return nil
}

func SetLogLevel(newLevel log.Level) {
	log.SetLevel(newLevel)
}

func CloseLogs() {
	logFile.Close()
}

func Init(newClientID string, newClientSecret string, logPath string) {
	SetClientID(newClientID)
	SetClientSecret(newClientSecret)
	err := InitLogs(logPath)
	if err != nil {
		fmt.Printf("couldn't init logs")
		os.Exit(1)
	}
}

func SetClientSecret(newClientSecret string) {
	clientSecret = newClientSecret
}

func SetClientID(newClientID string) {
	clientID = newClientID
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
	defer time.Sleep(time.Millisecond * 501)
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
		log.Infof("PATCH on %s with mark %d", urlTeam, mark)
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
		log.Infof("PATCH on %s with mark %d", urlProjectSession, mark)
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
	log.Debugf("calculating final mark for project-%s", parentID)
	for _, id := range childrenIDs {
		child, err := fetchProject(userID, id, accessToken)
		log.Debugf(`  fetched project-%d: {team_id: %d, validated?: %t, mark: %d}`, child.ProjectId, child.TeamId, child.Validated, child.FinalMark)
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

func CheckPoolProject(userID, moduleID string) {
	log.Infof("user-%s module-%s: starting checkup", userID, moduleID)
	err := UpdatePoolParentProject(userID, moduleID)
	if err != nil {
		fmt.Printf("user-%s module-%s: %s\n", userID, moduleID, err.Error())
		log.Errorf("user-%s module-%s: %s", userID, moduleID, err.Error())
		return
	}
	log.Infof("user-%s module-%s: everything went well", userID, moduleID)
}

func UpdatePoolParentProject(userID string, moduleID string) (err error) {
	// Load the configuration from the file
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("no clientid or clientsecret provided")
	}

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

	if parent.FinalMark >= mark {
		fmt.Printf("user already had a better grade\n")
		log.Infof("nothing updated: current mark for parent (%d) is higher or equal than new mark (%d)", parent.FinalMark, mark)
		return nil
	}
	err = updateProjectTeamMark(parent.TeamId, parent.Users[0].ProjectsUserId, mark, accessToken)
	if err != nil {
		return fmt.Errorf("error updating data: %w", err)
	}
	fmt.Printf("Succesfully updated team %d's mark to %d\n", parent.TeamId, mark)
	return nil
}
