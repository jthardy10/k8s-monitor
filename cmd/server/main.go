package main

import (
    "context"
    "log"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "k8s-monitor/internal/cluster"
    "k8s-monitor/internal/middleware"
    "k8s-monitor/internal/monitoring"
    "k8s-monitor/internal/storage"
)

var (
    registeredClusters map[string]cluster.ClusterRegistration
    store *storage.FileStorage
    monitoringService *monitoring.MonitoringService
)

func main() {
    store = storage.NewFileStorage("clusters.json")
    
    var err error
    registeredClusters, err = store.LoadClusters()
    if err != nil {
        log.Fatalf("Failed to load clusters: %v", err)
    }

    monitoringService = monitoring.NewMonitoringService(30 * time.Second)
    for _, reg := range registeredClusters {
        monitoringService.AddCluster(reg)
    }
    
    ctx := context.Background()
    monitoringService.Start(ctx)

    r := gin.Default()

    // Serve static files
    r.Static("/dashboard", "./static")
    
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
        })
    })
    
    api := r.Group("/clusters")
    api.Use(middleware.AuthRequired())
    {
        api.GET("", func(c *gin.Context) {
            var clusterNames []string
            for name := range registeredClusters {
                clusterNames = append(clusterNames, name)
            }
            c.JSON(http.StatusOK, gin.H{
                "clusters": clusterNames,
            })
        })

        api.POST("", func(c *gin.Context) {
            var newCluster cluster.ClusterRegistration
            if err := c.BindJSON(&newCluster); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
            }

            if err := newCluster.Validate(); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "validation failed",
                    "details": err.Error(),
                })
                return
            }

            registeredClusters[newCluster.Name] = newCluster
            monitoringService.AddCluster(newCluster)
            
            if err := store.SaveClusters(registeredClusters); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "failed to save cluster",
                    "details": err.Error(),
                })
                return
            }

            c.JSON(http.StatusCreated, gin.H{
                "message": "cluster registered successfully",
                "name": newCluster.Name,
                "environment": newCluster.Environment,
            })
        })

        api.GET("/:name", func(c *gin.Context) {
            clusterName := c.Param("name")
            
            metrics, err := monitoringService.GetMetrics(clusterName)
            if err != nil {
                c.JSON(http.StatusNotFound, gin.H{
                    "error": "cluster not found or no metrics available",
                    "name": clusterName,
                })
                return
            }

            c.JSON(http.StatusOK, gin.H{
                "name": clusterName,
                "environment": registeredClusters[clusterName].Environment,
                "metrics": metrics.Metrics,
                "lastUpdate": metrics.Timestamp,
                "status": "connected",
            })
        })

        api.GET("/:name/history", func(c *gin.Context) {
            clusterName := c.Param("name")
            
            history, err := monitoringService.GetMetricsHistory(clusterName)
            if err != nil {
                c.JSON(http.StatusNotFound, gin.H{
                    "error": "cluster not found",
                    "name": clusterName,
                })
                return
            }

            c.JSON(http.StatusOK, gin.H{
                "name": clusterName,
                "history": history,
            })
        })

        api.DELETE("/:name", func(c *gin.Context) {
            clusterName := c.Param("name")
            if _, exists := registeredClusters[clusterName]; !exists {
                c.JSON(http.StatusNotFound, gin.H{
                    "error": "cluster not found",
                    "name": clusterName,
                })
                return
            }

            delete(registeredClusters, clusterName)
            monitoringService.RemoveCluster(clusterName)
            
            if err := store.SaveClusters(registeredClusters); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "failed to save changes",
                    "details": err.Error(),
                })
                return
            }

            c.JSON(http.StatusOK, gin.H{
                "message": "cluster deleted successfully",
                "name": clusterName,
            })
        })
    }
    
    log.Fatal(r.Run(":8080"))
}
