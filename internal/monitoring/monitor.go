package monitoring

import (
    "context"
    "time"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

type Monitor struct {
    client *kubernetes.Clientset
}

func NewMonitor(client *kubernetes.Clientset) *Monitor {
    return &Monitor{client: client}
}

func (m *Monitor) GetPodStatus(namespace string) ([]string, error) {
    pods, err := m.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    
    var statuses []string
    for _, pod := range pods.Items {
        statuses = append(statuses, pod.Status.Phase)
    }
    return statuses, nil
}
