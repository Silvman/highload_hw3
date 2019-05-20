package server

import (
	"context"
	"flag"
	"highload_hw3/models"
	"highload_hw3/pkg/session"
	"highload_hw3/storage"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/go-playground/validator.v9"
)

type Server struct {
	httpSrv  *http.Server
	sm       session.SessionManagerClient
	users    storage.UserStorageI
	game     storage.GameStorageI
	validate *validator.Validate
	log      *log.Logger
}

func newServer(logger *log.Logger, sessionConn *grpc.ClientConn) *Server {
	return &Server{
		httpSrv: &http.Server{
			Addr:         config.Port,
			WriteTimeout: config.WriteTimeout,
			ReadTimeout:  config.ReadTimeout,
		},
		sm:       session.NewSessionManagerClient(sessionConn),
		users:    storage.GetUserStorage(),
		game:     storage.GetGameStorage(),
		validate: models.InitValidator(),
		log:      logger,
	}
}

func (srv *Server) createRoute() {
	r := mux.NewRouter()
	r.Use(srv.logginigMiddleware)
	r.Use(corsMiddleware)
	r.Use(srv.authMiddleware)

	// route for OPTIONS
	r.HandleFunc("/", srv.getSession).Methods(http.MethodOptions)

	userRouter := r.PathPrefix("/user").Subrouter()

	// GetScoreboard
	userRouter.HandleFunc("", srv.GetUsersWithOptions).Methods(http.MethodGet).
		Queries("limit", "{limit:[0-9]*?}", "offset", "{offset:[0-9]*?}")

	// Create/Get User
	userRouter.HandleFunc("", srv.createUser).Methods(http.MethodPost)
	userRouter.HandleFunc("/{id:[0-9]+}", srv.getUser).Methods(http.MethodGet)

	// UpdateUser
	userRouter.HandleFunc("/{id:[0-9]+}", srv.updateUser).Methods(http.MethodPatch)

	// GetAllUsers
	userRouter.HandleFunc("", srv.getUsers).Methods(http.MethodGet)

	sessionRouter := r.PathPrefix("/session").Subrouter()
	// Create/Get/Delete Session
	sessionRouter.HandleFunc("", srv.createSession).Methods(http.MethodPost)
	sessionRouter.HandleFunc("", srv.getSession).Methods(http.MethodGet)
	sessionRouter.HandleFunc("", srv.deleteSession).Methods(http.MethodDelete)

	gameRouter := r.PathPrefix("/game").Subrouter()
	// Handlers for work with game data
	gameRouter.Use(srv.authRequierMiddleware)
	gameRouter.HandleFunc("", srv.saveScore).Methods(http.MethodPost)
	gameRouter.HandleFunc("/shop", srv.getShopPositions).Methods(http.MethodGet)
	gameRouter.HandleFunc("/shop", srv.buyUpgrades).Methods(http.MethodPost)
	gameRouter.HandleFunc("/achievement", srv.getAchievement).Methods(http.MethodGet)

	srv.httpSrv.Handler = r
}

func StartApp() {
	flag.Parse()
	logger := log.New()
	logger.SetLevel(log.InfoLevel)
	// TODO add hook to Graylog
	logger.SetOutput(os.Stdout)
	err := storage.InitDB(config.DB)
	if err != nil {
		log.Warnln("can't init database", err.Error())
	}

	sessionConn, err := grpc.Dial(config.SessionAddr, grpc.WithInsecure())
	if err != nil {
		log.Warnln("can't connect to grpc")
		return
	}
	defer sessionConn.Close()
	srv := newServer(logger, sessionConn)
	srv.createRoute()
	go func() {
		logger.Infof("Starting server on %s", config.Port)
		if err := srv.httpSrv.ListenAndServe(); err != nil {
			log.Warnln(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	srv.httpSrv.Shutdown(ctx)
	log.Infoln("Shutdown server...")
	os.Exit(0)
}
