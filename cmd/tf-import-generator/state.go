package main

import (
	"fmt"
	"sort"
)

type States map[string]State

func (s States) Keys() []string {
	keys := make([]string, 0, len(s))

	for k := range s {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func (s States) CommonInstances() []string {
	count := make(map[string]int)

	for _, state := range s {
		for name := range state.MapInstances() {
			count[name]++
		}
	}

	var common []string

	for name, count := range count {
		if count == len(s) {
			common = append(common, name)
		}
	}

	sort.Strings(common)

	return common
}

func (s States) CommonResources() []string {
	count := make(map[string]int)

	for _, state := range s {
		for name := range state.MapResources() {
			count[name]++
		}
	}

	var common []string

	for name, count := range count {
		if count == len(s) {
			common = append(common, name)
		}
	}

	sort.Strings(common)

	return common
}

func (s States) MapResources(address string) map[string]*Resource {
	collated := make(map[string]*Resource)

	for key, state := range s {
		resources := state.MapResources()

		if resource := resources[address]; resource != nil {
			collated[key] = resource
		}
	}

	return collated
}

func (s States) MapInstances(address string) map[string]*Instance {
	collated := make(map[string]*Instance)

	for key, state := range s {
		instances := state.MapInstances()

		if resource := instances[address]; resource != nil {
			collated[key] = resource
		}
	}

	return collated
}

type State struct {
	Version   int       `json:"version"`
	Resources Resources `json:"resources"`
}

func (s *State) MapInstances() map[string]*Instance {
	instances := make(map[string]*Instance)

	for _, resource := range s.Resources {
		for _, instance := range resource.Instances {
			instances[resource.ID()+instance.Index()] = &instance
		}
	}

	return instances
}

func (s *State) MapResources() map[string]*Resource {
	resources := make(map[string]*Resource)

	for _, resource := range s.Resources {
		resources[resource.ID()] = &resource
	}

	return resources
}

type Resources []Resource

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

func (r Resource) Enumeration() string {
	switch {
	case len(r.Instances) > 0:
		first := r.Instances[0].Enumeration()
		for i := 1; i < len(r.Instances); i++ {
			if first != r.Instances[i].Enumeration() {
				return "mixed"
			}
		}
		return first
	default:
		return "none"
	}
}

type Instance map[string]interface{}

func (i Instance) Enumeration() string {
	for k, v := range i {
		if k == "index_key" {
			if _, ok := v.(float64); ok {
				return "count"
			}
			if _, ok := v.(string); ok {
				return "for_each"
			}
		}
	}

	return "native"
}

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

	switch {
	case resourceType == "time_rotating":
		return fmt.Sprintf(
			"%s,%s",
			attributes["rfc3339"].(string),
			attributes["rotation_rfc3339"].(string),
		)
	case resourceType == "aws_cloudwatch_event_target":
		return fmt.Sprintf(
			"%s/%s/%s",
			attributes["event_bus_name"].(string),
			attributes["rule"].(string),
			attributes["target_id"].(string),
		)

	case resourceType == "aws_iam_role_policy_attachment":
		return fmt.Sprintf(
			"%s/%s",
			attributes["role"].(string),
			attributes["policy_arn"].(string),
		)

	case resourceType == "kubernetes_manifest":

		if manifest, ok := attributes["manifest"].(map[string]interface{}); ok {
			if value, ok := manifest["value"].(map[string]interface{}); ok {
				if metadata, ok := value["metadata"].(map[string]interface{}); ok {
					return fmt.Sprintf(
						"apiVersion=%s,kind=%s,namespace=%s,name=%s",
						value["apiVersion"].(string),
						value["kind"].(string),
						metadata["namespace"].(string),
						metadata["name"].(string),
					)
				}
			}
		}

		return "?"

	case attributes["id"] != nil:
		return fmt.Sprintf(
			"%s",
			attributes["id"].(string),
		)

	default:
		fmt.Printf("Unknown resource type without an ID: %s\n", resourceType)
	}

	return "?"
}
