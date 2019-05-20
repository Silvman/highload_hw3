package session_server

import (
	"highload_hw3/pkg/session"
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type SessionManager struct {
	sync.Mutex
	sessions redis.Client
}

func NewSessionManager(conn *redis.Client) *SessionManager {
	return &SessionManager{
		Mutex:    sync.Mutex{},
		sessions: *conn,
	}
}

func (sm *SessionManager) Create(ctx context.Context, in *session.Session) (*session.SessionID, error) {
	log.Println("create session")
	ID, err := uuid.NewUUID()
	if err != nil {
		err = errors.Wrap(err, "can't create session-server ID")
		log.Println(err.Error())
		return nil, err
	}
	sessionID := session.SessionID{ID: ID.String()}
	dataSerialized, err := json.Marshal(in)
	if err != nil {
		err = errors.Wrap(err, "can't marshal session-server")
		log.Println(err.Error())
		return nil, err
	}
	mkey := "sessions:" + sessionID.ID
	sm.Lock()
	result := sm.sessions.Set(mkey, dataSerialized, time.Hour * 24 * 7)
	sm.Unlock()
	if err != nil {
		err = errors.Wrap(err, "can't insert valuer into redis")
		log.Println(err.Error())
		return nil, err
	}
	if msg, err := result.Result(); err != nil {
		err = errors.Wrap(err, "result from redis is not OK: " + msg )
		log.Println(err.Error())
		return nil, err
	}
	return &sessionID, nil
}

func (sm *SessionManager) Check(ctx context.Context, in *session.SessionID) (*session.Session, error) {
	log.Println("check session ", in.ID)
	mkey := "sessions:" + in.ID
	sm.Lock()
	result := sm.sessions.Get(mkey)
	sm.Unlock()
	if result.Err() != nil {
		return nil, errors.Wrap(result.Err(), "can't get data")
	}
	res, err := result.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "can't get data")
	}
	sess := &session.Session{}
	if err := sess.UnmarshalJSON(res); err != nil {
		err = errors.Wrap(err, "can't unpack session-server data")
		log.Println(err.Error())
		return nil, err
	}
	return sess, nil
}

func (sm *SessionManager) Delete(ctx context.Context, in *session.SessionID) (*session.Nothing, error) {
	log.Println("delete session")
	mkey := "sessions:" + in.ID
	sm.Lock()
	res := sm.sessions.Del(mkey)
	sm.Unlock()
	if res.Err() != nil {
		return nil, errors.Wrap(res.Err(), "can't del from redis")
	}
	return &session.Nothing{Dummy: true}, nil
}
