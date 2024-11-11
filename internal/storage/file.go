package storage

import (
    "encoding/json"
    "os"
    "sync"
    "k8s-monitor/internal/cluster"
)

type FileStorage struct {
    filepath string
    mu       sync.RWMutex
}

func NewFileStorage(filepath string) *FileStorage {
    return &FileStorage{
        filepath: filepath,
    }
}

func (f *FileStorage) SaveClusters(clusters map[string]cluster.ClusterRegistration) error {
    f.mu.Lock()
    defer f.mu.Unlock()

    data, err := json.Marshal(clusters)
    if err != nil {
        return err
    }

    return os.WriteFile(f.filepath, data, 0644)
}

func (f *FileStorage) LoadClusters() (map[string]cluster.ClusterRegistration, error) {
    f.mu.RLock()
    defer f.mu.RUnlock()

    data, err := os.ReadFile(f.filepath)
    if err != nil {
        if os.IsNotExist(err) {
            return make(map[string]cluster.ClusterRegistration), nil
        }
        return nil, err
    }

    var clusters map[string]cluster.ClusterRegistration
    if err := json.Unmarshal(data, &clusters); err != nil {
        return nil, err
    }

    return clusters, nil
}
