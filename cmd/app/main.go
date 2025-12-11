package main

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"

	"gosmol/internal/adapters/rest"
	"gosmol/internal/config"
	"gosmol/internal/service"
	"gosmol/internal/storage/psql"

	"gosmol/pkg/client/postgresql"
	"gosmol/pkg/logging"
)

const (
    envLocal = "local"
    envDev   = "dev"
    envProd  = "prod"
)

func main() {
  logging.Init()

  logger := logging.GetLogger()
  logger.Infoln("Logger enabled")

  jwtSecret := "my-secret-key"
  logger.Infof("secret %s", jwtSecret)

  logger.Infoln("Config initializing")

  router := httprouter.New()
  logger.Infoln("Router initializing")

  cfg := config.GetConfig()
  logger.Infof("DB CONFIG: Host=%s, Port=%d, Database=%s, Username=%s", 
    cfg.StorageConfig.Host, cfg.StorageConfig.Port, 
    cfg.StorageConfig.Database, cfg.StorageConfig.Username)
  logger.Infoln("Config initializing")

  postgreSQLClient, err := postgresql.NewClient(context.TODO(), 15, cfg.StorageConfig)
  if err != nil {
    logger.Fatalf("Failed to connect to database: %v", err)
  }

  logger.Infoln("Checking available databases...")
    rows, err := postgreSQLClient.Query(context.Background(), "SELECT datname FROM pg_database WHERE datistemplate = false;")
    if err == nil {
        defer rows.Close()
        for rows.Next() {
            var dbName string
            rows.Scan(&dbName)
            logger.Infof("Available database: %s", dbName)
        }
    }

    logger.Infoln("Checking tables in current database...")
    rows, err = postgreSQLClient.Query(context.Background(), 
        "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';")
    if err == nil {
        defer rows.Close()
        for rows.Next() {
            var tableName string
            rows.Scan(&tableName)
            logger.Infof("Table: %s", tableName)
        }
    }

    var actualUserCount int
    err = postgreSQLClient.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&actualUserCount)
    if err != nil {
        logger.Errorf("Count users failed: %v", err)
    } else {
        logger.Infof("Actual users in connected DB: %d", actualUserCount)
        
        rows, err := postgreSQLClient.Query(context.Background(), "SELECT id, email FROM users LIMIT 5")
        if err == nil {
            defer rows.Close()
            for rows.Next() {
                var id int64
                var email string
                rows.Scan(&id, &email)
                logger.Infof("   User: ID=%d, Email=%s", id, email)
            }
        }
    }

  logger.Infoln("Database initializing")

  twoFaRepo := psql.NewTwoFaRepo(postgreSQLClient)
  studentsRepo := psql.NewStudentsRepo(postgreSQLClient)
  studentsService := service.NewStudents(studentsRepo, twoFaRepo, jwtSecret)
  studentsHandler := rest.NewStudentsHandler(studentsService, logger)
  studentsHandler.Register(router, jwtSecret)

  diplomasRepo := psql.NewDiplomasRepo(postgreSQLClient)
  diplomasService := service.NewDiplomas(diplomasRepo)
  diplomasHandler := rest.NewDiplomasHandler(diplomasService, logger)
  diplomasHandler.Register(router, jwtSecret)

  logger.Infoln("📋 Registered routes:")
  router.HandleOPTIONS = true

  logger.Infof("Students routes: /api/auth/register, /api/auth/login, /api/auth/refresh")
  logger.Infof("Diplomas routes: /api/resources, /api/resource/:id")

  logger.Infoln("Students & diplomas initializing")

  cors := cors.New(cors.Options{
    AllowedOrigins:   []string{"*"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
    AllowedHeaders:   []string{"*"},
    ExposedHeaders:   []string{"*"},
    AllowCredentials: true,
    MaxAge:           12 * 3600,
  })

  handler := cors.Handler(router)
  logger.Infoln("Cors initializing")

  logger.Fatalln(http.ListenAndServe(":8888", handler))
}