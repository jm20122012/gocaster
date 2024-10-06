package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jm20122012/gocaster/internal/config"
	"github.com/jm20122012/gocaster/internal/session"
)

type Server struct {
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	logger   *slog.Logger
	conf     *config.Config
	upgrader websocket.Upgrader
}

func NewServer(ctx context.Context, cancel context.CancelFunc, logger *slog.Logger, conf *config.Config) (*Server, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return &Server{
		ctx:      ctx,
		cancel:   cancel,
		wg:       sync.WaitGroup{},
		logger:   logger,
		conf:     conf,
		upgrader: upgrader,
	}, nil
}

func (s *Server) Start() {
	// TODO: Add more endpoints for gathering server stats
	// TODO: Maybe use Chi for http server routes
	http.HandleFunc("/ws", s.httpHandleFunc)

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%s", s.conf.ServerPort),
	}

	// Start http server as go routine so we can catch the ctx.Done() signal
	go func() {
		s.logger.Info("Starting HTTP server", "serverPort", s.conf.ServerPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("error starting http server", "error", err)
		}
	}()

	<-s.ctx.Done()
	s.logger.Info("Global context cancelled - shutting down")

	httpCtx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpCancel()

	err := httpServer.Shutdown(httpCtx)
	if err != nil {
		s.logger.Error("error shutting down http server", "error", err)
	}

	s.logger.Info("HTTP server shutdown successful")
}

func (s *Server) httpHandleFunc(w http.ResponseWriter, r *http.Request) {
	clientID := uuid.NewString()

	s.logger.Debug("New client connected", "clientID", clientID)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("error upgrading client connection", "clientID", clientID, "error", err)
		return
	}

	go s.registerClient(clientID, conn)
}

func (s *Server) registerClient(clientID string, conn *websocket.Conn) {

	sess, err := session.NewSession(
		s.ctx,
		s.logger,
		conn,
		clientID,
	)
	if err != nil {
		s.logger.Error("error registering client", "clientID", clientID)
	}

	go sess.StartClientHandler()
}
