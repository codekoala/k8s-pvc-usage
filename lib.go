package pvcusage

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	// NodeList represents the list of nodes in the k8s cluster.
	NodeList struct {
		Items []Node `json:"items"`
	}

	// Node represents an individual k8s node.
	Node struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	}

	// NodePodsList represents the list of pods scheduled on a k8s node.
	NodePodsList struct {
		Pods []Pod `json:"pods"`
	}

	// Pod represents a k8s Pod workload.
	Pod struct {
		Volumes []Volume `json:"volume"`
	}

	// Volume represents information about a volume attached to a Pod.
	Volume struct {
		Time           *time.Time `json:"time"`
		AvailableBytes float64    `json:"availableBytes"`
		CapacityBytes  float64    `json:"capacityBytes"`
		UsedBytes      float64    `json:"usedBytes"`
		InodesFree     float64    `json:"inodesFree"`
		Inodes         float64    `json:"inodes"`
		InodesUsed     float64    `json:"inodesUsed"`
		Name           string     `json:"name"`
		PvcRef         *PvcRef    `json:"pvcRef"`
	}

	// PvcRef represents a relationship between a Volume and a PVC.
	PvcRef struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}

	PvcStats struct {
		Name           string  `json:"name"`
		Namespace      string  `json:"namespace"`
		AvailableBytes float64 `json:"availableBytes"`
		CapacityBytes  float64 `json:"capacityBytes"`
		UsedBytes      float64 `json:"usedBytes"`
	}
)

func GetNodes(api *Client) []Node {
	return GetNodesCtx(context.Background(), api)
}

func GetPvcUsage(api *Client) []PvcStats {
	return GetPvcUsageCtx(context.Background(), api)
}

func GetNodePvcUsage(api *Client, node Node) []PvcStats {
	return GetNodePvcUsageCtx(context.Background(), api, node)
}

func GetNodesCtx(ctx context.Context, api *Client) []Node {
	var nodes NodeList

	resp := api.Req(ctx, "nodes").
		Send().Scan(&nodes)
	if err := resp.Error(); err != nil {
		log.Error().
			Err(err).
			Msg("error listing nodes")
	}

	return nodes.Items
}

func GetPvcUsageCtx(ctx context.Context, api *Client) (stats []PvcStats) {
	for _, node := range GetNodesCtx(ctx, api) {
		stats = append(stats, GetNodePvcUsageCtx(ctx, api, node)...)
	}

	return stats
}

func GetNodePvcUsageCtx(ctx context.Context, api *Client, node Node) (stats []PvcStats) {
	var pods NodePodsList

	resp := api.Req(ctx, "nodes", node.Metadata.Name, "proxy", "stats", "summary").
		Send().Scan(&pods)
	if err := resp.Error(); err != nil {
		log.Error().
			Err(err).
			Str("node", node.Metadata.Name).
			Msg("error getting node stats")
		return
	}

	for _, pod := range pods.Pods {
		for _, vol := range pod.Volumes {
			if vol.PvcRef == nil || vol.CapacityBytes <= 0 {
				continue
			}

			stats = append(stats, PvcStats{
				Name:           vol.PvcRef.Name,
				Namespace:      vol.PvcRef.Namespace,
				AvailableBytes: vol.AvailableBytes,
				CapacityBytes:  vol.CapacityBytes,
				UsedBytes:      vol.UsedBytes,
			})
		}
	}

	return stats
}

func (pvc PvcStats) Usage() float64 {
	if pvc.CapacityBytes <= 0 {
		return 0
	}

	return 100.0 * pvc.UsedBytes / pvc.CapacityBytes
}
