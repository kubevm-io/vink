package watcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher/filter"
	"github.com/kubevm.io/vink/pkg/watcher/utils"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

// func Watch[T any](ctx context.Context, kubeInformerFactory informer.KubeInformerFactory, sink EventSink[T], namespacedName *types.NamespacedName) error {
// 	batched := NewBatchedSink(sink, 500*time.Millisecond, 10)
// 	defer batched.Stop()

// 	var (
// 		zero   T
// 		gvr, _ = dynamicx.ResolveGVRAndGVK(zero)

// 		initialSynced = make(map[string]struct{})
// 		mu            sync.Mutex
// 	)

// 	informer, ok := kubeInformerFactory.InformerForGVR(gvr)
// 	if !ok {
// 		return fmt.Errorf("failed to find informer for %s", gvr.String())
// 	}

// 	if err := sink.OnReady(); err != nil {
// 		return fmt.Errorf("OnReady failed: %w", err)
// 	}

// 	filterFuncs := make([]filter.FilterFunc, 0)
// 	filterFuncs = append(filterFuncs, filter.FilterFuncWithNamespacedName(namespacedName))

// 	objs := informer.GetIndexer().List()
// 	for _, obj := range objs {
// 		value, err := dynamicx.FromObject[T](obj)
// 		if err != nil {
// 			return fmt.Errorf("convert failed: %w", err)
// 		}

// 		pass, err := filter.PassesAllFilters(filterFuncs, obj)
// 		if err != nil {
// 			return err
// 		}
// 		if !pass {
// 			continue
// 		}

// 		metaobj, err := utils.GetMetaObject(obj)
// 		if err != nil {
// 			return err
// 		}
// 		initialSynced[string(metaobj.GetUID())] = struct{}{}
// 		if err := batched.OnAdd(value); err != nil {
// 			return err
// 		}
// 	}
// 	batched.buffer.flush()

// 	var errCh = make(chan error)

// 	eventHandler := cache.ResourceEventHandlerFuncs{
// 		AddFunc: func(obj any) {
// 			pass, err := filter.PassesAllFilters(filterFuncs, obj)
// 			if err != nil {
// 				errCh <- fmt.Errorf("AddFunc: failed to PassesAllFilters, error: %w", err)
// 				return
// 			}
// 			if !pass {
// 				return
// 			}

// 			value, err := dynamicx.FromObject[T](obj)
// 			if err != nil {
// 				errCh <- fmt.Errorf("AddFunc: convert failed: %w", err)
// 				return
// 			}

// 			metaobj, err := utils.GetMetaObject(obj)
// 			if err != nil {
// 				errCh <- fmt.Errorf("AddFunc: get meta object failed: %w", err)
// 				return
// 			}

// 			mu.Lock()
// 			if _, exists := initialSynced[string(metaobj.GetUID())]; exists {
// 				delete(initialSynced, string(metaobj.GetUID()))
// 				mu.Unlock()
// 				return
// 			}
// 			mu.Unlock()

// 			if err := batched.OnAdd(value); err != nil {
// 				errCh <- fmt.Errorf("OnAdd failed: %w", err)
// 				return
// 			}
// 		},
// 		UpdateFunc: func(_, newObj any) {
// 			pass, err := filter.PassesAllFilters(filterFuncs, newObj)
// 			if err != nil {
// 				errCh <- fmt.Errorf("UpdateFunc: failed to PassesAllFilters, error: %w", err)
// 				return
// 			}
// 			if !pass {
// 				return
// 			}

// 			value, err := dynamicx.FromObject[T](newObj)
// 			if err != nil {
// 				errCh <- fmt.Errorf("UpdateFunc: convert failed: %w", err)
// 				return
// 			}
// 			if err := sink.OnUpdate(value); err != nil {
// 				errCh <- fmt.Errorf("OnUpdate failed: %w", err)
// 				return
// 			}
// 		},
// 		DeleteFunc: func(obj any) {
// 			pass, err := filter.PassesAllFilters(filterFuncs, obj)
// 			if err != nil {
// 				errCh <- fmt.Errorf("DeleteFunc: failed to PassesAllFilters, error: %w", err)
// 				return
// 			}
// 			if !pass {
// 				return
// 			}

// 			value, err := dynamicx.FromObject[T](obj)
// 			if err != nil {
// 				errCh <- fmt.Errorf("DeleteFunc: convert failed: %w", err)
// 				return
// 			}
// 			if err := sink.OnDelete(value); err != nil {
// 				errCh <- fmt.Errorf("OnDelete failed: %w", err)
// 				return
// 			}
// 		},
// 	}

// 	registration, err := informer.AddEventHandler(eventHandler)
// 	if err != nil {
// 		return fmt.Errorf("failed to AddEventHandler, error: %v", err)
// 	}
// 	defer informer.RemoveEventHandler(registration)

// 	select {
// 	case err := <-errCh:
// 		fmt.Println(err)
// 		return err
// 	case <-ctx.Done():
// 		fmt.Println("stopping resource watch")
// 		return nil
// 	}
// }

func Watch[T any](ctx context.Context, kubeInformerFactory informer.KubeInformerFactory, sink EventSink[T], namespacedName *types.NamespacedName) error {
	batched := NewBatchedSink(sink, 500*time.Millisecond, 10)
	defer batched.Stop()

	var (
		zero          T
		gvr, _        = dynamicx.ResolveGVRAndGVK(zero)
		initialSynced = make(map[string]struct{})
		mu            sync.Mutex
	)

	informer, ok := kubeInformerFactory.InformerForGVR(gvr)
	if !ok {
		return fmt.Errorf("failed to find informer for %s", gvr.String())
	}

	if err := sink.OnReady(); err != nil {
		return fmt.Errorf("OnReady failed: %w", err)
	}

	filterFuncs := []filter.FilterFunc{filter.FilterFuncWithNamespacedName(namespacedName)}

	objs := informer.GetIndexer().List()
	for _, obj := range objs {
		value, err := dynamicx.FromObject[T](obj)
		if err != nil {
			return fmt.Errorf("convert failed: %w", err)
		}

		pass, err := filter.PassesAllFilters(filterFuncs, obj)
		if err != nil {
			return err
		}
		if !pass {
			continue
		}

		metaobj, err := utils.GetMetaObject(obj)
		if err != nil {
			return err
		}
		initialSynced[string(metaobj.GetUID())] = struct{}{}
		if err := batched.OnAdd(value); err != nil {
			return err
		}
	}
	batched.Flush()

	var errCh = make(chan error)

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			handleEvent(obj, eventAdd, batched, initialSynced, &mu, filterFuncs, errCh)
		},
		UpdateFunc: func(_, newObj any) {
			handleEvent(newObj, eventUpdate, batched, nil, &mu, filterFuncs, errCh)
		},
		DeleteFunc: func(obj any) {
			handleEvent(obj, eventDelete, batched, nil, &mu, filterFuncs, errCh)
		},
	}

	registration, err := informer.AddEventHandler(eventHandler)
	if err != nil {
		return fmt.Errorf("failed to AddEventHandler: %w", err)
	}
	defer informer.RemoveEventHandler(registration)

	select {
	case err := <-errCh:
		fmt.Println(err)
		return err
	case <-ctx.Done():
		fmt.Println("stopping resource watch")
		return nil
	}
}

func handleEvent[T any](obj any, typ eventType, bs *BatchedSink[T], initial map[string]struct{}, mu *sync.Mutex, filters []filter.FilterFunc, errCh chan error) {
	pass, err := filter.PassesAllFilters(filters, obj)
	if err != nil {
		errCh <- fmt.Errorf("failed to PassesAllFilters: %w", err)
		return
	}
	if !pass {
		return
	}

	value, err := dynamicx.FromObject[T](obj)
	if err != nil {
		errCh <- fmt.Errorf("convert failed: %w", err)
		return
	}

	if typ == eventAdd && initial != nil {
		metaobj, err := utils.GetMetaObject(obj)
		if err != nil {
			errCh <- fmt.Errorf("get meta object failed: %w", err)
			return
		}
		mu.Lock()
		if _, exists := initial[string(metaobj.GetUID())]; exists {
			delete(initial, string(metaobj.GetUID()))
			mu.Unlock()
			return
		}
		mu.Unlock()
	}

	switch typ {
	case eventAdd:
		_ = bs.OnAdd(value)
	case eventUpdate:
		_ = bs.OnUpdate(value)
	case eventDelete:
		_ = bs.OnDelete(value)
	}
}

type eventType int

const (
	eventAdd eventType = iota
	eventUpdate
	eventDelete
)

type queuedEvent[T any] struct {
	typ eventType
	obj T
}

type EventSink[T any] interface {
	OnAdd(obj T) error
	OnUpdate(obj T) error
	OnDelete(obj T) error
	OnReady() error
}

type BatchedSink[T any] struct {
	sink       EventSink[T]
	window     time.Duration
	buffer     *batchedBuffer[T]
	stopSignal chan struct{}
}

func NewBatchedSink[T any](sink EventSink[T], window time.Duration, threshold int) *BatchedSink[T] {
	bs := &BatchedSink[T]{
		sink:       sink,
		window:     window,
		stopSignal: make(chan struct{}),
	}

	bs.buffer = newBatchedBuffer(func(events []queuedEvent[T]) {
		for _, e := range events {
			switch e.typ {
			case eventAdd:
				_ = sink.OnAdd(e.obj)
			case eventUpdate:
				_ = sink.OnUpdate(e.obj)
			case eventDelete:
				_ = sink.OnDelete(e.obj)
			}
		}
	}, threshold)

	go bs.loop()
	return bs
}

func (bs *BatchedSink[T]) loop() {
	ticker := time.NewTicker(bs.window)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			bs.buffer.flush()
		case <-bs.stopSignal:
			return
		}
	}
}

func (bs *BatchedSink[T]) Stop() {
	close(bs.stopSignal)
}

func (bs *BatchedSink[T]) OnAdd(obj T) error {
	bs.buffer.push(eventAdd, obj)
	return nil
}

func (bs *BatchedSink[T]) OnUpdate(obj T) error {
	bs.buffer.push(eventUpdate, obj)
	return nil
}

func (bs *BatchedSink[T]) OnDelete(obj T) error {
	bs.buffer.push(eventDelete, obj)
	return nil
}

func (bs *BatchedSink[T]) OnReady() error {
	return bs.sink.OnReady()
}

func (bs *BatchedSink[T]) Flush() {
	bs.buffer.flush()
}

type batchedBuffer[T any] struct {
	mu        sync.Mutex
	events    []queuedEvent[T]
	flushFn   func([]queuedEvent[T])
	threshold int
}

func newBatchedBuffer[T any](flushFn func([]queuedEvent[T]), threshold int) *batchedBuffer[T] {
	return &batchedBuffer[T]{
		flushFn:   flushFn,
		threshold: threshold,
	}
}

func (b *batchedBuffer[T]) push(typ eventType, obj T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, queuedEvent[T]{typ: typ, obj: obj})
	if len(b.events) >= b.threshold {
		b.flushLocked()
	}
}

func (b *batchedBuffer[T]) flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushLocked()
}

func (b *batchedBuffer[T]) flushLocked() {
	if len(b.events) == 0 {
		return
	}
	b.flushFn(b.events)
	b.events = nil
}

// func NewBatchedSink[T any](sink EventSink[T], window time.Duration, threshold int) *BatchedSink[T] {
// 	bs := &BatchedSink[T]{
// 		sink:       sink,
// 		window:     window,
// 		stopSignal: make(chan struct{}),
// 	}
// 	bs.buffer = newBatchedBuffer(func(added, updated, deleted []T) {
// 		if len(added) > 0 {
// 			for _, a := range added {
// 				_ = sink.OnAdd(a)
// 			}
// 		}
// 		if len(updated) > 0 {
// 			for _, u := range updated {
// 				_ = sink.OnUpdate(u)
// 			}
// 		}
// 		if len(deleted) > 0 {
// 			for _, d := range deleted {
// 				_ = sink.OnDelete(d)
// 			}
// 		}
// 	}, threshold)

// 	go bs.loop()
// 	return bs
// }

// type EventSink[T any] interface {
// 	OnAdd(obj T) error

// 	OnUpdate(obj T) error

// 	OnDelete(obj T) error

// 	OnReady() error
// }

// type BatchedSink[T any] struct {
// 	sink       EventSink[T]
// 	window     time.Duration
// 	buffer     *batchedBuffer[T]
// 	stopSignal chan struct{}
// }

// func (bs *BatchedSink[T]) loop() {
// 	ticker := time.NewTicker(bs.window)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-ticker.C:
// 			bs.buffer.flush()
// 		case <-bs.stopSignal:
// 			return
// 		}
// 	}
// }

// func (bs *BatchedSink[T]) Stop() {
// 	close(bs.stopSignal)
// }

// func (bs *BatchedSink[T]) OnAdd(obj T) error {
// 	bs.buffer.push(&obj, nil, nil)
// 	return nil
// }

// func (bs *BatchedSink[T]) OnUpdate(obj T) error {
// 	bs.buffer.push(nil, &obj, nil)
// 	return nil
// }

// func (bs *BatchedSink[T]) OnDelete(obj T) error {
// 	bs.buffer.push(nil, nil, &obj)
// 	return nil
// }

// func (bs *BatchedSink[T]) OnReady() error {
// 	return bs.sink.OnReady()
// }

// type batchedBuffer[T any] struct {
// 	mu        sync.Mutex
// 	added     []T
// 	updated   []T
// 	deleted   []T
// 	flushFn   func([]T, []T, []T)
// 	threshold int
// }

// func newBatchedBuffer[T any](flushFn func([]T, []T, []T), threshold int) *batchedBuffer[T] {
// 	return &batchedBuffer[T]{
// 		flushFn:   flushFn,
// 		threshold: threshold,
// 	}
// }

// func (b *batchedBuffer[T]) push(add, update, del *T) {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()

// 	if add != nil {
// 		b.added = append(b.added, *add)
// 	}
// 	if update != nil {
// 		b.updated = append(b.updated, *update)
// 	}
// 	if del != nil {
// 		b.deleted = append(b.deleted, *del)
// 	}

// 	total := len(b.added) + len(b.updated) + len(b.deleted)
// 	if total >= b.threshold {
// 		b.flushLocked()
// 	}
// }

// func (b *batchedBuffer[T]) flush() {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	b.flushLocked()
// }

// func (b *batchedBuffer[T]) flushLocked() {
// 	if len(b.added) == 0 && len(b.updated) == 0 && len(b.deleted) == 0 {
// 		return
// 	}
// 	b.flushFn(b.added, b.updated, b.deleted)
// 	b.added, b.updated, b.deleted = nil, nil, nil
// }
