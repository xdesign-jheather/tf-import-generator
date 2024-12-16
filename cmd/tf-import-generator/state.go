package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type State struct {
	Version   int        `json:"version"`
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Module    string     `json:"module"`
	Mode      string     `json:"mode"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Instances []Instance `json:"instances"`
}

func (r Resource) ID() string {
	switch {
	case r.Module == "":
		return fmt.Sprintf("%s.%s", r.Type, r.Name)
	default:
		return fmt.Sprintf("%s.%s.%s", r.Module, r.Type, r.Name)
	}
}

type Instance map[string]interface{}

func (i Instance) Index() string {
	for k, v := range i {
		if k == "index_key" {
			if fl, ok := v.(float64); ok {
				return fmt.Sprintf("[%.0f]", fl)
			}
			if str, ok := v.(string); ok {
				return fmt.Sprintf(`["%s"]`, str)
			}
		}
	}

	return ""
}

func (i Instance) Import(resourceType string) string {
	attributes := i["attributes"].(map[string]interface{})

	switch resourceType {
	case "aws_cloudwatch_event_target":
		return fmt.Sprintf(
			"%s/%s/%s",
			attributes["event_bus_name"].(string),
			attributes["rule"].(string),
			attributes["target_id"].(string),
		)

	case "aws_iam_role_policy_attachment":
		return fmt.Sprintf(
			"%s/%s",
			attributes["role"].(string),
			attributes["policy_arn"].(string),
		)
	default:
		return fmt.Sprintf(
			"%s",
			attributes["id"].(string),
		)
	}
}

func readState(r io.Reader) State {
	var state State

	if err := json.NewDecoder(r).Decode(&state); err != nil {
		log.Fatal(err)
	}

	return state
}
