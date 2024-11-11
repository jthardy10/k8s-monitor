package cluster

import (
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

type Manager struct {
    clusters map[string]*kubernetes.Clientset
}

func NewManager() *Manager {
    return &Manager{
        clusters: make(map[string]*kubernetes.Clientset),
    }
}

func (m *Manager) RegisterCluster(name string, config *rest.Config) error {
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return err
    }
    m.clusters[name] = clientset
    return nil
}
