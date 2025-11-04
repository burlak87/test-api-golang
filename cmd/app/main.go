package main

import (
	"context"
	"net/http"
	"os"

	"github.com/joho/godotenv"
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

  err := godotenv.Load()
  if err != nil {
    logger.Fatalf("Error load file .env: %v", err)
  }

  jwtSecret := os.Getenv("JWT_SECRET")
  logger.Infof("secret %s", jwtSecret)

  dbURL := os.Getenv("SUPABASE_CONNECT_DB")
  logger.Infof("db url %s", dbURL)

  logger.Infoln("Config initializing")

  router := httprouter.New()
  logger.Infoln("Router initializing")

  cfg := config.GetConfig()
  logger.Infoln("Config initializing")

  postgreSQLClient, err := postgresql.NewClient(context.TODO(), 3, cfg.StorageConfig)
  if err != nil {
    logger.Fatalf("%v", err)
  }

  logger.Infoln("Database initializing")
    
  studentsRepo := psql.NewStudentsRepo(postgreSQLClient)
  studentsService := service.NewStudents(studentsRepo, jwtSecret)
  studentsHandler := rest.NewStudentsHandler(studentsService, logger)
  studentsHandler.Register(router)

  diplomasRepo := psql.NewDiplomasRepo(postgreSQLClient)
  diplomasService := service.NewDiplomas(diplomasRepo)
  diplomasHandler := rest.NewDiplomasHandler(diplomasService, logger)
  diplomasHandler.Register(router, jwtSecret)

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
  // log.Fatal(http.ListenAndServe(":8888", handler))
}

// func start(router *httprouter.Router) {
//   listener, err := net.Listen("tcp", ":1234")
//   if err != nil {
//     panic(err)
//   }

//   server := &http.Server{
//     Handler: router,
//     WriteTimeout: 15 * time.Second,
//     ReadTimeout: 15 * time.Second,
//   }

//   log.Fatalln(server.Serve(listener))
// }