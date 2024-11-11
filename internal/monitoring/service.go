package monitoring

import (
    "context"
    "fmt"
    "sync"
    "time"
    "k8s-monitor/internal/cluster"
)

type MetricsHistory struct {
    Timestamp time.Time
    Metrics   *cluster.ClusterMetrics
}

type ClusterMonitor struct {
    clusterReg      cluster.ClusterRegistration
    metricsHistory  []MetricsHistory
    lastUpdate      time.Time
    mu             sync.RWMutex
    maxHistory     int
}

type MonitoringService struct {
    monitors  map[string]*ClusterMonitor
    mu       sync.RWMutex
    interval time.Duration
}

func NewMonitoringService(interval time.Duration) *MonitoringService {
    return &MonitoringService{
        monitors:  make(map[string]*ClusterMonitor),
        interval: interval,
    }
}

func (s *MonitoringService) AddCluster(reg cluster.ClusterRegistration) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.monitors[reg.Name] = &ClusterMonitor{
        clusterReg:  reg,
        maxHistory: 100, // Keep last 100 metrics
    }
}

func (s *MonitoringService) RemoveCluster(name string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.monitors, name)
}

func (s *MonitoringService) GetMetrics(name string) (*MetricsHistory, error) {
    s.mu.RLock()
    monitor, exists := s.monitors[name]
    s.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("cluster not found")
    }

    monitor.mu.RLock()
    defer monitor.mu.RUnlock()
    
    if len(monitor.metricsHistory) == 0 {
        return nil, fmt.Errorf("no metrics available")
    }
    
    return &monitor.metricsHistory[len(monitor.metricsHistory)-1], nil
}

func (s *MonitoringService) GetMetricsHistory(name string) ([]MetricsHistory, error) {
    s.mu.RLock()
    monitor, exists := s.monitors[name]
    s.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("cluster not found")
    }

    monitor.mu.RLock()
    defer monitor.mu.RUnlock()
    
    history := make([]MetricsHistory, len(monitor.metricsHistory))
    copy(history, monitor.metricsHistory)
    return history, nil
}

func (s *MonitoringService) Start(ctx context.Context) {
    ticker := time.NewTicker(s.interval)
    go func() {
        for {
            select {
            case <-ctx.Done():
                ticker.Stop()
                return
            case <-ticker.C:
                s.updateMetrics()
            }
        }
    }()
}

func (s *MonitoringService) updateMetrics() {
    s.mu.RLock()
    clusters := make(map[string]*ClusterMonitor)
    for k, v := range s.monitors {
        clusters[k] = v
    }
    s.mu.RUnlock()

    for _, monitor := range clusters {
        monitor.mu.Lock()
        metrics, err := cluster.GetClusterMetrics(monitor.clusterReg.KubeConfig)
        if err == nil {
            history := MetricsHistory{
                Timestamp: time.Now(),
                Metrics:   metrics,
            }
            
            monitor.metricsHistory = append(monitor.metricsHistory, history)
            if len(monitor.metricsHistory) > monitor.maxHistory {
                monitor.metricsHistory = monitor.metricsHistory[1:]
            }
            monitor.lastUpdate = time.Now()
        }
        monitor.mu.Unlock()
    }
}
