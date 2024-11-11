package cluster

import (
    "context"
    "encoding/base64"
    "fmt"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodStatus struct {
    Running   int `json:"running"`
    Pending   int `json:"pending"`
    Failed    int `json:"failed"`
    Succeeded int `json:"succeeded"`
}

type ResourceMetrics struct {
    CPU    string `json:"cpu"`
    Memory string `json:"memory"`
}

type NodeMetrics struct {
    Name      string         `json:"name"`
    Resources ResourceMetrics `json:"resources"`
    Capacity  ResourceMetrics `json:"capacity"`
}

type ClusterMetrics struct {
    Nodes struct {
        Total    int           `json:"total"`
        Ready    int           `json:"ready"`
        NotReady int           `json:"notReady"`
        Metrics  []NodeMetrics `json:"metrics"`
    } `json:"nodes"`
    Pods struct {
        Total       int               `json:"total"`
        Status      PodStatus         `json:"status"`
        ByNamespace map[string]int    `json:"byNamespace"`
        Resources   map[string]ResourceMetrics `json:"resources"`
    } `json:"pods"`
    Services struct {
        Total   int            `json:"total"`
        ByType  map[string]int `json:"byType"`
    } `json:"services"`
    Namespaces int `json:"namespaces"`
    ClusterResources struct {
        TotalCPU    string `json:"totalCPU"`
        TotalMemory string `json:"totalMemory"`
        UsedCPU     string `json:"usedCPU"`
        UsedMemory  string `json:"usedMemory"`
    } `json:"clusterResources"`
}

func GetClusterMetrics(kubeconfig string) (*ClusterMetrics, error) {
    kubeconfigBytes, err := base64.StdEncoding.DecodeString(kubeconfig)
    if err != nil {
        return nil, fmt.Errorf("invalid kubeconfig encoding: %v", err)
    }

    config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
    if err != nil {
        return nil, err
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    // Create metrics client
    metricsClient, err := versioned.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    ctx := context.Background()
    metrics := &ClusterMetrics{}
    metrics.Pods.ByNamespace = make(map[string]int)
    metrics.Services.ByType = make(map[string]int)
    metrics.Pods.Resources = make(map[string]ResourceMetrics)
    
    // Get Nodes and their metrics
    nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    nodeMetrics, _ := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
    nodeMetricsMap := make(map[string]ResourceMetrics)
    if nodeMetrics != nil {
        for _, nm := range nodeMetrics.Items {
            nodeMetricsMap[nm.Name] = ResourceMetrics{
                CPU:    nm.Usage.Cpu().String(),
                Memory: nm.Usage.Memory().String(),
            }
        }
    }

    metrics.Nodes.Total = len(nodes.Items)
    var totalCPU, totalMemory int64
    for _, node := range nodes.Items {
        isReady := false
        for _, condition := range node.Status.Conditions {
            if condition.Type == "Ready" {
                if condition.Status == "True" {
                    isReady = true
                }
                break
            }
        }
        if isReady {
            metrics.Nodes.Ready++
        } else {
            metrics.Nodes.NotReady++
        }

        nodeMetric := NodeMetrics{
            Name: node.Name,
            Capacity: ResourceMetrics{
                CPU:    node.Status.Capacity.Cpu().String(),
                Memory: node.Status.Capacity.Memory().String(),
            },
        }
        if m, exists := nodeMetricsMap[node.Name]; exists {
            nodeMetric.Resources = m
        }
        metrics.Nodes.Metrics = append(metrics.Nodes.Metrics, nodeMetric)

        totalCPU += node.Status.Capacity.Cpu().Value()
        totalMemory += node.Status.Capacity.Memory().Value()
    }

    metrics.ClusterResources.TotalCPU = fmt.Sprintf("%d cores", totalCPU)
    metrics.ClusterResources.TotalMemory = fmt.Sprintf("%dMi", totalMemory/(1024*1024))

    // Get Pods
    pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }

    podMetrics, _ := metricsClient.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
    podMetricsMap := make(map[string]ResourceMetrics)
    if podMetrics != nil {
        for _, pm := range podMetrics.Items {
            var cpuTotal, memoryTotal int64
            for _, container := range pm.Containers {
                cpuTotal += container.Usage.Cpu().Value()
                memoryTotal += container.Usage.Memory().Value()
            }
            podMetricsMap[pm.Namespace+"/"+pm.Name] = ResourceMetrics{
                CPU:    fmt.Sprintf("%dm", cpuTotal),
                Memory: fmt.Sprintf("%dMi", memoryTotal/(1024*1024)),
            }
        }
    }

    metrics.Pods.Total = len(pods.Items)
    var usedCPU, usedMemory int64
    for _, pod := range pods.Items {
        // Count by namespace
        metrics.Pods.ByNamespace[pod.Namespace]++
        
        // Count by status
        switch pod.Status.Phase {
        case "Running":
            metrics.Pods.Status.Running++
        case "Pending":
            metrics.Pods.Status.Pending++
        case "Failed":
            metrics.Pods.Status.Failed++
        case "Succeeded":
            metrics.Pods.Status.Succeeded++
        }

        // Add resource metrics
        if m, exists := podMetricsMap[pod.Namespace+"/"+pod.Name]; exists {
            metrics.Pods.Resources[pod.Name] = m
        }
    }

    metrics.ClusterResources.UsedCPU = fmt.Sprintf("%dm", usedCPU)
    metrics.ClusterResources.UsedMemory = fmt.Sprintf("%dMi", usedMemory/(1024*1024))

    // Get Services
    services, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    metrics.Services.Total = len(services.Items)
    for _, svc := range services.Items {
        metrics.Services.ByType[string(svc.Spec.Type)]++
    }

    // Get Namespaces
    namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    metrics.Namespaces = len(namespaces.Items)

    return metrics, nil
}
