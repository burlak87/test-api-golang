package app

import (
	"gosmol/internal/adapters/rest"
	"gosmol/internal/service"
	"gosmol/internal/storage/psql"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"

	"gosmol/pkg/client/supabase"
	"gosmol/pkg/logging"
)

const (
    envLocal = "local"
    envDev   = "dev"
    envProd  = "prod"
)

func main() {
  logger := logging.GetLogger()
  logger.Infoln("Logger enabled")

  jwtSecret := os.Getenv("JWT_SECRET")
  logger.Infoln("Config initializing")

  router := httprouter.New()
  logger.Infoln("Router initializing")

  supabaseClient, err := supabase.NewSupabaseClient()
  if err != nil {
    logger.Fatalln("Failed to connect to Supabase", err)
    // log.Fatal("Failed to connect to Supabase:", err)
  }
  defer supabaseClient.Close()
  logger.Infoln("Database initializing")
    
  studentsRepo := psql.NewStudentsRepo(supabaseClient.Pool)
  studentsService := service.NewStudents(studentsRepo, jwtSecret)
  studentsHandler := rest.NewStudentsHandler(studentsService, logger)
  studentsHandler.Register(router)

  diplomasRepo := psql.NewDiplomasRepo(supabaseClient.Pool)
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