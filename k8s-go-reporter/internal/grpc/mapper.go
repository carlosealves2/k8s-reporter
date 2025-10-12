package grpc

import (
	pb "github.com/carlosealves2/k8s-go-reporter/internal/gen/k8sreporter/v1"
	"github.com/carlosealves2/k8s-go-reporter/internal/services"
)

// PodToProto converts a domain Pod to a protobuf Pod
// This follows the Single Responsibility Principle - dedicated conversion logic
func PodToProto(pod services.Pod) *pb.Pod {
	return &pb.Pod{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    pod.Status,
		NodeName:  pod.NodeName,
	}
}

// PodsToProto converts a slice of domain Pods to protobuf Pods
// This is a convenience function for bulk conversions
func PodsToProto(pods []services.Pod) []*pb.Pod {
	pbPods := make([]*pb.Pod, 0, len(pods))
	for _, pod := range pods {
		pbPods = append(pbPods, PodToProto(pod))
	}
	return pbPods
}
