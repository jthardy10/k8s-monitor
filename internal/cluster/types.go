package cluster

import (
    "encoding/base64"
    "fmt"
    "k8s.io/client-go/tools/clientcmd"
    "strings"
)

type ClusterRegistration struct {
    Name        string `json:"name"`
    KubeConfig  string `json:"kubeConfig"`
    Environment string `json:"environment"`
}

func (c *ClusterRegistration) Validate() error {
    if strings.TrimSpace(c.Name) == "" {
        return fmt.Errorf("cluster name is required")
    }

    if strings.TrimSpace(c.Environment) == "" {
        return fmt.Errorf("environment is required")
    }

    // Decode base64 kubeconfig
    kubeconfigBytes, err := base64.StdEncoding.DecodeString(c.KubeConfig)
    if err != nil {
        return fmt.Errorf("invalid kubeconfig encoding: %v", err)
    }

    // Validate kubeconfig
    _, err = clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
    if err != nil {
        return fmt.Errorf("invalid kubeconfig: %v", err)
    }

    return nil
}

func (c *ClusterRegistration) GetKubeConfigBytes() ([]byte, error) {
    return base64.StdEncoding.DecodeString(c.KubeConfig)
}
