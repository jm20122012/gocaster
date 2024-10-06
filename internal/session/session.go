package session

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type ISession interface {
	clientReader(sessionCtx context.Context, sessionCancel context.CancelFunc)
	clientWriter(sessionCtx context.Context, sessionCancel context.CancelFunc)
}

type Session struct {
	ctx           context.Context
	sessionCtx    context.Context
	sessionCancel context.CancelFunc
	logger        *slog.Logger
	conn          *websocket.Conn
	clientID      string
	wg            sync.WaitGroup
	rxChan        chan []byte // Channel for messages received from the client
	txChan        chan []byte // Channel for messages sent to the client
	mutex         sync.RWMutex
}

func NewSession(
	ctx context.Context,
	logger *slog.Logger,
	conn *websocket.Conn,
	clientID string,
) (*Session, error) {

	s, c := context.WithCancel(ctx)
	return &Session{
		ctx:           ctx,
		sessionCtx:    s,
		sessionCancel: c,
		logger:        logger,
		conn:          conn,
		clientID:      clientID,
		wg:            sync.WaitGroup{},
		rxChan:        make(chan []byte, 2048),
		txChan:        make(chan []byte, 2048),
		mutex:         sync.RWMutex{},
	}, nil
}

func (s *Session) StartClientHandler() {
	s.logger.Debug("Starting client handler", "clientID", s.clientID)

	// ! Make sure to pass the correct number of goroutines to Add()
	s.wg.Add(4)

	go s.contextMonitor()
	go s.handleClientMsgs()
	go s.clientReader()
	go s.clientWriter()

	<-s.ctx.Done() // Wait for server context to finish

	s.sessionCancel()

	s.conn.Close()

	s.wg.Wait()

	s.logger.Info("session exiting", "clientID", s.clientID)
}

func (s *Session) contextMonitor() {
	defer func() {
		s.logger.Debug("exiting client context monitor", "clientID", s.clientID)
		s.wg.Done()
	}()

	<-s.ctx.Done()
	s.sessionCancel()
	s.conn.Close()
}

func (s *Session) clientReader() {
	defer func() {
		s.logger.Debug("exiting client reader", "clientID", s.clientID)
		s.wg.Done()
	}()

	s.logger.Debug("starting client reader", "clientID", s.clientID)

	for {
		select {
		case <-s.sessionCtx.Done():
			return
		default:
			_, msg, err := s.conn.ReadMessage()
			if err != nil {
				// If the error isn't a code in this list, IsUnexpectedError will return false
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
					s.logger.Debug("client disconnected", "clientID", s.clientID, "error", err)
				} else {
					s.logger.Error("error reading message from client", "clientID", s.clientID, "error", err)
				}
				s.sessionCancel()
				return
			}

			s.logger.Debug("received message from client", "clientID", s.clientID, "msg", string(msg))
			s.rxChan <- msg
		}
	}
}

func (s *Session) handleClientMsgs() {
	defer func() {
		s.logger.Debug("exiting client message handler", "clientID", s.clientID)
		s.wg.Done()
	}()

	for {
		select {
		case <-s.sessionCtx.Done():
			return
		case msg := <-s.rxChan:
			s.logger.Debug("Parsing client message", "clientID", s.clientID, "msg", msg)
		}
	}
}

func (s *Session) clientWriter() {
	defer func() {
		s.logger.Debug("exiting client writer", "clientID", s.clientID)
		s.wg.Done()
	}()

	s.logger.Debug("starting client writer", "clientID", s.clientID)

	for {
		select {
		case <-s.sessionCtx.Done():
			return
			// TODO: add client writer functionality
			// default:
		}
	}
}
