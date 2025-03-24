package poolmanager

import (
	"encoding/json"
	"io/ioutil"
	"slices"
)

// Struct to represent each parent with their children
type PoolProject struct {
	ParentName  string   `json:"parent_name"`
	ParentID    string   `json:"parent_id"`
	ChildrenIDs []string `json:"children_ids"`
}

var Pools []PoolProject

// Function to read and parse the config file
func LoadConfig(filename string) error {
	// Read the JSON file
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Unmarshal the JSON into the config slice
	err = json.Unmarshal(file, &Pools)
	if err != nil {
		return err
	}
	return nil
}

func GetProjectParentID(childrenID string) (parentID string) {
	for _, parent := range Pools {
		if slices.Contains(parent.ChildrenIDs, childrenID) {
			return parent.ParentID
		}
	}
	return ""
}

func GetProjectChildrenIDs(parentID string) (childrenIDs []string) {
	for _, parent := range Pools {
		if parent.ParentID == parentID {
			return parent.ChildrenIDs
		}
	}
	return nil
}
