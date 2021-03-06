package kvstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"xorkevin.dev/governor"
)

const (
	kvpathSeparator = ":"
)

type (
	// Resulter returns the result of a string command in a multi after it has executed
	Resulter interface {
		Result() (string, error)
	}

	// IntResulter returns the result of an int command in a multi after it has executed
	IntResulter interface {
		Result() (int64, error)
	}

	// Multi is a kvstore multi
	Multi interface {
		Get(key string) Resulter
		GetInt(key string) IntResulter
		Set(key, val string, seconds int64)
		Del(key ...string)
		Incr(key string, delta int64) IntResulter
		Expire(key string, seconds int64)
		Subkey(keypath ...string) string
		Subtree(prefix string) Multi
		Exec() error
	}

	// KVStore is a service wrapper around a kv store client
	KVStore interface {
		Get(key string) (string, error)
		GetInt(key string) (int64, error)
		Set(key, val string, seconds int64) error
		Del(key ...string) error
		Incr(key string, delta int64) (int64, error)
		Expire(key string, seconds int64) error
		Subkey(keypath ...string) string
		Multi() (Multi, error)
		Tx() (Multi, error)
		Subtree(prefix string) KVStore
	}

	// Service is a KVStore and governor.Service
	Service interface {
		governor.Service
		KVStore
	}

	getClientRes struct {
		client *redis.Client
		err    error
	}

	getOp struct {
		res chan<- getClientRes
	}

	service struct {
		client     *redis.Client
		auth       string
		addr       string
		dbname     int
		config     governor.SecretReader
		logger     governor.Logger
		ops        chan getOp
		ready      bool
		hbfailed   int
		hbinterval int
		hbmaxfail  int
		done       <-chan struct{}
	}

	ctxKeyRootKV struct{}

	ctxKeyKVStore struct{}
)

// getCtxRootKV returns a root KVStore from the context
func getCtxRootKV(inj governor.Injector) KVStore {
	v := inj.Get(ctxKeyRootKV{})
	if v == nil {
		return nil
	}
	return v.(KVStore)
}

// setCtxRootKV sets a root KVStore in the context
func setCtxRootKV(inj governor.Injector, k KVStore) {
	inj.Set(ctxKeyRootKV{}, k)
}

// GetCtxKVStore returns a KVStore from the context
func GetCtxKVStore(inj governor.Injector) KVStore {
	v := inj.Get(ctxKeyKVStore{})
	if v == nil {
		return nil
	}
	return v.(KVStore)
}

// setCtxKVStore sets a KVStore in the context
func setCtxKVStore(inj governor.Injector, k KVStore) {
	inj.Set(ctxKeyKVStore{}, k)
}

// NewSubtreeInCtx creates a new kv subtree with a prefix and sets it in the context
func NewSubtreeInCtx(inj governor.Injector, prefix string) {
	kv := getCtxRootKV(inj)
	setCtxKVStore(inj, kv.Subtree(prefix))
}

// New creates a new cache service
func New() Service {
	return &service{
		ops:      make(chan getOp),
		ready:    false,
		hbfailed: 0,
	}
}

func (s *service) Register(inj governor.Injector, r governor.ConfigRegistrar, jr governor.JobRegistrar) {
	setCtxRootKV(inj, s)

	r.SetDefault("auth", "")
	r.SetDefault("dbname", 0)
	r.SetDefault("host", "localhost")
	r.SetDefault("port", "6379")
	r.SetDefault("hbinterval", 5)
	r.SetDefault("hbmaxfail", 5)
}

type (
	// ErrConn is returned on a kvstore connection error
	ErrConn struct{}
	// ErrClient is returned for unknown client errors
	ErrClient struct{}
	// ErrNotFound is returned when a key is not found
	ErrNotFound struct{}
	// ErrVal is returned for invalid value errors
	ErrVal struct{}
)

func (e ErrConn) Error() string {
	return "KVStore connection error"
}

func (e ErrClient) Error() string {
	return "KVStore client error"
}

func (e ErrNotFound) Error() string {
	return "Key not found"
}

func (e ErrVal) Error() string {
	return "Invalid value"
}

func (s *service) Init(ctx context.Context, c governor.Config, r governor.ConfigReader, l governor.Logger, m governor.Router) error {
	s.logger = l
	l = s.logger.WithData(map[string]string{
		"phase": "init",
	})

	s.config = r

	s.addr = fmt.Sprintf("%s:%s", r.GetStr("host"), r.GetStr("port"))
	s.dbname = r.GetInt("dbname")
	s.hbinterval = r.GetInt("hbinterval")
	s.hbmaxfail = r.GetInt("hbmaxfail")

	l.Info("loaded config", map[string]string{
		"addr":       s.addr,
		"dbname":     strconv.Itoa(s.dbname),
		"hbinterval": strconv.Itoa(s.hbinterval),
		"hbmaxfail":  strconv.Itoa(s.hbmaxfail),
	})

	done := make(chan struct{})
	go s.execute(ctx, done)
	s.done = done

	if _, err := s.getClient(); err != nil {
		return err
	}
	return nil
}

func (s *service) execute(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(time.Duration(s.hbinterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.closeClient()
			return
		case <-ticker.C:
			s.handlePing()
		case op := <-s.ops:
			client, err := s.handleGetClient()
			op.res <- getClientRes{
				client: client,
				err:    err,
			}
			close(op.res)
		}
	}
}

func (s *service) handlePing() {
	if s.client != nil {
		_, err := s.client.Ping().Result()
		if err == nil {
			s.ready = true
			s.hbfailed = 0
			return
		}
		s.hbfailed++
		if s.hbfailed < s.hbmaxfail {
			s.logger.Warn("failed to ping kvstore", map[string]string{
				"error":      err.Error(),
				"actiontype": "pingkv",
				"address":    s.addr,
				"dbname":     strconv.Itoa(s.dbname),
			})
			return
		}
		s.logger.Error("failed max pings to kvstore", map[string]string{
			"error":      err.Error(),
			"actiontype": "pingkvmax",
			"address":    s.addr,
			"dbname":     strconv.Itoa(s.dbname),
		})
		s.ready = false
		s.hbfailed = 0
		s.auth = ""
		s.config.InvalidateSecret("auth")
	}
	if _, err := s.handleGetClient(); err != nil {
		s.logger.Error("failed to create kvstore client", map[string]string{
			"error":      err.Error(),
			"actiontype": "createkvclient",
		})
	}
}

func (s *service) handleGetClient() (*redis.Client, error) {
	authsecret, err := s.config.GetSecret("auth")
	if err != nil {
		return nil, err
	}
	auth, ok := authsecret["password"].(string)
	if !ok || auth == "" {
		return nil, governor.ErrWithKind(nil, governor.ErrInvalidConfig{}, "Invalid secret")
	}
	if auth == s.auth {
		return s.client, nil
	}

	s.closeClient()

	client := redis.NewClient(&redis.Options{
		Addr:     s.addr,
		Password: auth,
		DB:       s.dbname,
	})
	if _, err := client.Ping().Result(); err != nil {
		s.config.InvalidateSecret("auth")
		return nil, governor.ErrWithKind(err, ErrConn{}, "Failed to ping kvstore")
	}

	s.client = client
	s.auth = auth
	s.ready = true
	s.hbfailed = 0
	s.logger.Info(fmt.Sprintf("established connection to %s dbname %d", s.addr, s.dbname), nil)
	return s.client, nil
}

func (s *service) closeClient() {
	if s.client == nil {
		return
	}
	if err := s.client.Close(); err != nil {
		s.logger.Error("failed to close kvstore connection", map[string]string{
			"error":      err.Error(),
			"actiontype": "closekverr",
			"address":    s.addr,
			"dbname":     strconv.Itoa(s.dbname),
		})
	} else {
		s.logger.Info("closed kvstore connection", map[string]string{
			"actiontype": "closekvok",
			"address":    s.addr,
			"dbname":     strconv.Itoa(s.dbname),
		})
	}
	s.client = nil
	s.auth = ""
}

func (s *service) Setup(req governor.ReqSetup) error {
	return nil
}

func (s *service) PostSetup(req governor.ReqSetup) error {
	return nil
}

func (s *service) Start(ctx context.Context) error {
	return nil
}

func (s *service) Stop(ctx context.Context) {
	l := s.logger.WithData(map[string]string{
		"phase": "stop",
	})
	select {
	case <-s.done:
		return
	case <-ctx.Done():
		l.Warn("failed to stop", nil)
	}
}

func (s *service) Health() error {
	if !s.ready {
		return governor.ErrWithKind(nil, ErrConn{}, "KVStore service not ready")
	}
	return nil
}

func (s *service) getClient() (*redis.Client, error) {
	res := make(chan getClientRes)
	op := getOp{
		res: res,
	}
	select {
	case <-s.done:
		return nil, governor.ErrWithKind(nil, ErrConn{}, "KVStore service shutdown")
	case s.ops <- op:
		v := <-res
		return v.client, v.err
	}
}

func (s *service) Get(key string) (string, error) {
	client, err := s.getClient()
	if err != nil {
		return "", err
	}
	val, err := client.Get(key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", governor.ErrWithKind(err, ErrNotFound{}, "Key not found")
		}
		return "", governor.ErrWithKind(err, ErrClient{}, "Failed to get key")
	}
	return val, nil
}

func (s *service) GetInt(key string) (int64, error) {
	client, err := s.getClient()
	if err != nil {
		return 0, err
	}
	val, err := client.Get(key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, governor.ErrWithKind(err, ErrNotFound{}, "Key not found")
		}
		return 0, governor.ErrWithKind(err, ErrClient{}, "Failed to get key")
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, governor.ErrWithKind(err, ErrVal{}, "Invalid int value")
	}
	return num, nil
}

func (s *service) Set(key, val string, seconds int64) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	if err := client.Set(key, val, time.Duration(seconds)*time.Second).Err(); err != nil {
		return governor.ErrWithKind(err, ErrClient{}, "Failed to set key")
	}
	return nil
}

func (s *service) Del(key ...string) error {
	if len(key) == 0 {
		return nil
	}

	client, err := s.getClient()
	if err != nil {
		return err
	}

	if err := client.Del(key...).Err(); err != nil {
		return governor.ErrWithKind(err, ErrClient{}, "Failed to delete key")
	}
	return nil
}

func (s *service) Incr(key string, delta int64) (int64, error) {
	client, err := s.getClient()
	if err != nil {
		return 0, err
	}
	val, err := client.IncrBy(key, delta).Result()
	if err != nil {
		return 0, governor.ErrWithKind(err, ErrClient{}, "Failed to incr key")
	}
	return val, nil
}

func (s *service) Expire(key string, seconds int64) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	if err := client.Expire(key, time.Duration(seconds)*time.Second).Err(); err != nil {
		return governor.ErrWithKind(err, ErrClient{}, "Failed to set expire key")
	}
	return nil
}

func (s *service) Subkey(keypath ...string) string {
	if len(keypath) == 0 {
		return ""
	}
	return strings.Join(keypath, kvpathSeparator)
}

func (s *service) Multi() (Multi, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	return &baseMulti{
		base: client.Pipeline(),
	}, nil
}

func (s *service) Tx() (Multi, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	return &baseMulti{
		base: client.TxPipeline(),
	}, nil
}

func (s *service) Subtree(prefix string) KVStore {
	return &tree{
		prefix: prefix,
		base:   s,
	}
}

type (
	baseMulti struct {
		base redis.Pipeliner
	}

	multi struct {
		prefix string
		base   *baseMulti
	}
)

func (t *baseMulti) Get(key string) Resulter {
	return &resulter{
		res: t.base.Get(key),
	}
}

func (t *baseMulti) GetInt(key string) IntResulter {
	return &intResulter{
		res: t.base.Get(key),
	}
}

func (t *baseMulti) Set(key, val string, seconds int64) {
	t.base.Set(key, val, time.Duration(seconds)*time.Second)
}

func (t *baseMulti) Del(key ...string) {
	t.base.Del(key...)
}

func (t *baseMulti) Incr(key string, delta int64) IntResulter {
	return &intCmdResulter{
		res: t.base.IncrBy(key, delta),
	}
}

func (t *baseMulti) Expire(key string, seconds int64) {
	t.base.Expire(key, time.Duration(seconds)*time.Second)
}

func (t *baseMulti) Subkey(keypath ...string) string {
	if len(keypath) == 0 {
		return ""
	}
	return strings.Join(keypath, kvpathSeparator)
}

func (t *baseMulti) Subtree(prefix string) Multi {
	return &multi{
		prefix: prefix,
		base:   t,
	}
}

func (t *baseMulti) Exec() error {
	if _, err := t.base.Exec(); err != nil {
		if !errors.Is(err, redis.Nil) {
			return governor.ErrWithKind(err, ErrNotFound{}, "Failed to execute multi")
		}
	}
	return nil
}

func (t *multi) Get(key string) Resulter {
	return t.base.Get(t.prefix + kvpathSeparator + key)
}

func (t *multi) GetInt(key string) IntResulter {
	return t.base.GetInt(t.prefix + kvpathSeparator + key)
}

func (t *multi) Set(key, val string, seconds int64) {
	t.base.Set(t.prefix+kvpathSeparator+key, val, seconds)
}

func (t *multi) Del(key ...string) {
	args := make([]string, 0, len(key))
	for _, i := range key {
		args = append(args, t.prefix+kvpathSeparator+i)
	}
	t.base.Del(args...)
}

func (t *multi) Incr(key string, delta int64) IntResulter {
	return t.base.Incr(t.prefix+kvpathSeparator+key, delta)
}

func (t *multi) Expire(key string, seconds int64) {
	t.base.Expire(key, seconds)
}

func (t *multi) Exec() error {
	return t.base.Exec()
}

func (t *multi) Subkey(keypath ...string) string {
	if len(keypath) == 0 {
		return ""
	}
	return strings.Join(keypath, kvpathSeparator)
}

func (t *multi) Subtree(prefix string) Multi {
	return &multi{
		prefix: t.prefix + kvpathSeparator + prefix,
		base:   t.base,
	}
}

type (
	tree struct {
		prefix string
		base   *service
	}
)

func (t *tree) Get(key string) (string, error) {
	return t.base.Get(t.prefix + kvpathSeparator + key)
}

func (t *tree) GetInt(key string) (int64, error) {
	return t.base.GetInt(t.prefix + kvpathSeparator + key)
}

func (t *tree) Set(key, val string, seconds int64) error {
	return t.base.Set(t.prefix+kvpathSeparator+key, val, seconds)
}

func (t *tree) Del(key ...string) error {
	args := make([]string, 0, len(key))
	for _, i := range key {
		args = append(args, t.prefix+kvpathSeparator+i)
	}
	return t.base.Del(args...)
}

func (t *tree) Incr(key string, delta int64) (int64, error) {
	return t.base.Incr(t.prefix+kvpathSeparator+key, delta)
}

func (t *tree) Expire(key string, seconds int64) error {
	return t.base.Expire(t.prefix+kvpathSeparator+key, seconds)
}

func (t *tree) Subkey(keypath ...string) string {
	if len(keypath) == 0 {
		return ""
	}
	return strings.Join(keypath, kvpathSeparator)
}

func (t *tree) Multi() (Multi, error) {
	tx, err := t.base.Multi()
	if err != nil {
		return nil, err
	}
	return tx.Subtree(t.prefix), nil
}

func (t *tree) Tx() (Multi, error) {
	tx, err := t.base.Tx()
	if err != nil {
		return nil, err
	}
	return tx.Subtree(t.prefix), nil
}

func (t *tree) Subtree(prefix string) KVStore {
	return &tree{
		prefix: t.prefix + kvpathSeparator + prefix,
		base:   t.base,
	}
}

type (
	resulter struct {
		res *redis.StringCmd
	}
)

func (r *resulter) Result() (string, error) {
	val, err := r.res.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", governor.ErrWithKind(err, ErrNotFound{}, "Key not found")
		}
		return "", governor.ErrWithKind(err, ErrClient{}, "Failed to get key")
	}
	return val, nil
}

type (
	intCmdResulter struct {
		res *redis.IntCmd
	}
)

func (r *intCmdResulter) Result() (int64, error) {
	val, err := r.res.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, governor.ErrWithKind(err, ErrNotFound{}, "Key not found")
		}
		return 0, governor.ErrWithKind(err, ErrClient{}, "Failed to get key")
	}
	return val, nil
}

type (
	intResulter struct {
		res *redis.StringCmd
	}
)

func (r *intResulter) Result() (int64, error) {
	val, err := r.res.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, governor.ErrWithKind(err, ErrNotFound{}, "Key not found")
		}
		return 0, governor.ErrWithKind(err, ErrClient{}, "Failed to get key")
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, governor.ErrWithKind(err, ErrVal{}, "Invalid int value")
	}
	return num, nil
}
