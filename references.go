package gorm_fixtures

import (
	"context"
	"errors"
	"log"
	"sync"
)

const ReferenceCtxKey = "_reference_"

var ErrReferenceNotFound = errors.New("reference not found")

type LoadCtx struct {
	ctx        context.Context
	references map[string]interface{}
	mu         *sync.RWMutex
}

func NewLoadCtx(ctx context.Context) *LoadCtx {
	return &LoadCtx{ctx: ctx, references: make(map[string]interface{}), mu: &sync.RWMutex{}}
}

func (l *LoadCtx) Context() context.Context {
	return l.ctx
}

func (l *LoadCtx) MustGetReference(id string) interface{} {
	v, err := l.GetReference(id)
	if err != nil {
		log.Fatalf("reference %s not found", id)
	}
	return v
}

func (l *LoadCtx) GetReference(id string) (interface{}, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	val, exists := l.references[id]
	if !exists {
		return nil, ErrReferenceNotFound
	}

	return val, nil
}

func (l *LoadCtx) SetReference(id string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.references[id] = value
}
