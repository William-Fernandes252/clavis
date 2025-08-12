package keys

import (
	"container/heap"
	"encoding/gob"
	"sync"
	"time"

	"github.com/William-Fernandes252/clavis/internal/data"
	"github.com/William-Fernandes252/clavis/internal/errors"
)

// Expiration represents a point in time when a key expires.
// It is a wrapper around time.Time to provide additional methods
// for checking expiration status and calculating time to live (TTL).
type Expiration interface {
	// IsNever returns true if the expiration means "never expire".
	IsNever() bool

	// IsExpired returns true if the expiration time is in the past.
	IsExpired() bool

	// IsFuture returns true if the expiration is a valid future time.
	IsFuture() bool

	// TTL returns the remaining time until expiration, or 0 if IsNever().
	// A negative value means the key is already expired.
	TTL() time.Duration

	// Unix returns the expiration time as a UNIX timestamp in seconds.
	Unix() int64

	// Time returns the time.Time that represents the expiration time.
	Time() time.Time
}

// expiration is a concrete implementation of the Expiration interface.
// It wraps a time.Time value and provides methods to check expiration status.
// It is used to represent the expiration time of keys in the system.
type expiration time.Time

// Never is the zero (no expiration) value
var Never Expiration = expiration(time.Time{})

// FromDuration returns an Expiration that expires after the given duration.
func FromDuration(d time.Duration) Expiration {
	if d <= 0 {
		return Never
	}
	return expiration(time.Now().Add(d))
}

// FromSeconds returns an Expiration that expires N seconds from now.
func FromSeconds(seconds int64) Expiration {
	return FromDuration(time.Duration(seconds) * time.Second)
}

// FromMilliseconds returns an Expiration that expires N milliseconds from now.
func FromMilliseconds(ms int64) Expiration {
	return FromDuration(time.Duration(ms) * time.Millisecond)
}

// FromUnix creates an Expiration for the given UNIX timestamp (seconds).
func FromUnix(unixSeconds int64) Expiration {
	return expiration(time.Unix(unixSeconds, 0))
}

// FromUnixMilli creates an Expiration from UNIX time in milliseconds.
func FromUnixMilli(ms int64) Expiration {
	sec := ms / 1000
	nsec := (ms % 1000) * int64(time.Millisecond)
	return expiration(time.Unix(sec, nsec))
}

// Until creates an Expiration for a specific time.
func Until(t time.Time) Expiration {
	return expiration(t)
}

// IsNever returns true if the expiration means "never expire".
func (e expiration) IsNever() bool {
	return time.Time(e).IsZero()
}

// IsExpired returns true if the expiration time is in the past.
func (e expiration) IsExpired() bool {
	if e.IsNever() {
		return false
	}
	return time.Now().After(time.Time(e))
}

// IsValid returns true if the expiration is a valid future time.
func (e expiration) IsFuture() bool {
	if e.IsNever() {
		return true // Never expiration is always valid
	}
	t := time.Time(e)
	return !t.IsZero() && t.After(time.Now())
}

// TTL returns the remaining time until expiration, or 0 if IsNever().
// A negative value means the key is already expired.
func (e expiration) TTL() time.Duration {
	if e.IsNever() {
		return 0
	}
	return time.Until(time.Time(e))
}

// Unix returns the expiration time as a UNIX timestamp in seconds.
func (e expiration) Unix() int64 {
	if e.IsNever() {
		return 0
	}
	return time.Time(e).Unix()
}

// String returns a string representation of the expiration time.
// If the expiration is "never", it returns "never".
// Otherwise, it returns the time in a human-readable format.
func (e expiration) String() string {
	if e.IsNever() {
		return "never"
	}
	return time.Time(e).String()
}

// Time returns the time.Time that represents the expiration time.
func (e expiration) Time() time.Time {
	return time.Time(e)
}

// expirationCodec implements the data.Codec interface for Expiration.
// It is used internally to convert Expiration values to and from byte slices for storage.
type expirationCodec struct{}

// expiration implements the data.Deserializer interface for Expiration.
func newExpirationCodec() *expirationCodec {
	return &expirationCodec{}
}

// Serialize implements the data.Serializer interface for expiration.
// It is used internally to convert the expiration to a byte slice for storage.
func (e *expirationCodec) Serialize(expiration Expiration) ([]byte, errors.Error) {
	raw, err := expiration.Time().MarshalBinary()
	if err != nil {
		return nil, data.NewDataError("serialization-failed", "Failed to serialize expiration", err)
	}
	return raw, nil
}

// Deserialize implements the data.Deserializer interface for expiration.
// It is used internally to convert a byte slice back into an expiration.
func (e *expirationCodec) Deserialize(raw []byte) (Expiration, errors.Error) {
	if len(raw) == 0 {
		return Never, nil // No data means no expiration
	}
	var t time.Time
	if err := t.UnmarshalBinary(raw); err != nil {
		return nil, data.NewDataError("deserialization-failed", "Failed to deserialize expiration", err)
	}
	value := expiration(t)
	return value, nil
}

// Type returns the type name ("expiration") for the expiration codec.
func (e *expirationCodec) Type() data.Type {
	return data.Type("expiration")
}

// ExpireFunc is a function type that expires a key.
// It is called by the ExpirationManager when a key expires.
// The function should handle the deletion logic, such as removing the key from storage, logging, or notifying other components.
// It is expected to be
//   - thread-safe and able to handle concurrent calls, as multiple keys may expire at the same time.
//   - idempotent, meaning calling it multiple times for an already expired key should have no adverse effects.
//   - efficient, as it may be called frequently in a high-throughput system.
//   - able to handle errors gracefully, such as logging or retrying if necessary.
type ExpireFunc func(key Key)

// ExpirationManager handles scheduling and canceling expirations.
type ExpirationManager interface {
	// Schedule key to expire at the given absolute time.
	Schedule(key Key, expiresAt Expiration) errors.Error

	// Cancel prevents any pending expiration for key.
	Cancel(key Key)

	// Expiration gets the expiration for a key.
	Get(key Key) (Expiration, errors.Error)

	// Start begins the manager?s background work.
	// The expire function will be called when a key expires.
	// It should handle the deletion logic, such as removing the key from storage.
	Start(expire ExpireFunc)

	// Stop the manager?s background work.
	Stop()
}

type keyExpirationScheduling struct {
	key       Key
	expiresAt Expiration
	index     int // index in the priority queue
}

// keyExpirationPriorityQueue implements heap.Interface over []*scheduledKeyEntry
type keyExpirationPriorityQueue []*keyExpirationScheduling

func (sq keyExpirationPriorityQueue) Len() int {
	return len(sq)
}

func (sq keyExpirationPriorityQueue) Less(i, j int) bool {
	return sq[i].expiresAt.Time().Before(sq[j].expiresAt.Time())
}

func (sq keyExpirationPriorityQueue) Swap(i, j int) {
	sq[i], sq[j] = sq[j], sq[i]
}

func (sq *keyExpirationPriorityQueue) Push(x any) {
	e := x.(*keyExpirationScheduling)
	e.index = len(*sq) // Set index to the end of the slice
	*sq = append(*sq, e)
}

func (sq *keyExpirationPriorityQueue) Pop() any {
	old := *sq
	n := len(old)
	e := old[n-1]
	*sq = old[:n-1]
	return e
}

// PriorityQueueExpirationManager evicts keys via a priority queue
type PriorityQueueExpirationManager struct {
	mu      sync.Mutex
	heap    keyExpirationPriorityQueue
	entries map[string]*keyExpirationScheduling
	wakeUp  chan struct{}
	stop    chan struct{}
}

func NewPriorityQueueExpirationManager() *PriorityQueueExpirationManager {
	m := &PriorityQueueExpirationManager{
		heap:    make(keyExpirationPriorityQueue, 0),
		entries: make(map[string]*keyExpirationScheduling),
		wakeUp:  make(chan struct{}),
		stop:    make(chan struct{}),
	}
	heap.Init(&m.heap)
	return m
}

// Schedule or reschedule key?s expiration
func (m *PriorityQueueExpirationManager) Schedule(key Key, expiresAt Expiration) errors.Error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.entries[key.String()]; ok {
		e.expiresAt = expiresAt
		heap.Fix(&m.heap, e.index)
	} else {
		e := &keyExpirationScheduling{key: key, expiresAt: expiresAt}
		heap.Push(&m.heap, e)
		m.entries[key.String()] = e
	}

	// wake up loop so it can adjust its timer
	select {
	case m.wakeUp <- struct{}{}:
	default:
	}

	return nil
}

// Get implements ExpirationManager.
func (m *PriorityQueueExpirationManager) Get(key Key) (Expiration, errors.Error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.entries[key.String()]; ok {
		return e.expiresAt, nil
	}
	return nil, NewStorageError("key-not-found", "Key not found", nil)
}

// Cancel removes a pending expiration (e.g. key deleted/reset)
func (m *PriorityQueueExpirationManager) Cancel(key Key) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.entries[key.String()]; ok {
		heap.Remove(&m.heap, e.index)
		delete(m.entries, key.String())
	}
}

// Start waits until the next scheduledEntry?s deadline and fires deletes
func (m *PriorityQueueExpirationManager) Start(expire ExpireFunc) {
	for {
		m.mu.Lock()
		if len(m.heap) == 0 {
			m.mu.Unlock()
			select {
			case <-m.wakeUp:
				continue
			case <-m.stop:
				return
			}
		}
		next := m.heap[0].expiresAt.Time()
		now := time.Now()
		delay := next.Sub(now)
		m.mu.Unlock()

		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-m.wakeUp:
				continue
			case <-m.stop:
				return
			}
		}

		// fire all expired entries
		m.mu.Lock()
		for len(m.heap) > 0 && !time.Now().Before(m.heap[0].expiresAt.Time()) {
			e := heap.Pop(&m.heap).(*keyExpirationScheduling)
			delete(m.entries, e.key.String())
			go expire(e.key)
		}
		m.mu.Unlock()
	}
}

// Stop the manager
func (m *PriorityQueueExpirationManager) Stop() {
	close(m.stop)
}

var _ ExpirationManager = (*PriorityQueueExpirationManager)(nil)

// TimingWheelExpirationManager implements a timing wheel for key expiration.
type TimingWheelExpirationManager struct {
	interval time.Duration // slot duration (e.g. 1s)
	slots    [][]Key       // circular buffer of keys
	size     int           // number of slots
	current  int           // current slot index
	mu       sync.Mutex
	stop     chan struct{}
}

// NewTimingWheel builds a wheel with slotInterval and wheelSize
func NewTimingWheelExpirationManager(slotInterval time.Duration, wheelSize int) *TimingWheelExpirationManager {
	tw := &TimingWheelExpirationManager{
		interval: slotInterval,
		size:     wheelSize,
		slots:    make([][]Key, wheelSize),
		stop:     make(chan struct{}),
	}
	return tw
}

// Schedule a key to expire at the given time.
func (tw *TimingWheelExpirationManager) Schedule(key Key, expiresAt Expiration) errors.Error {
	ticks := int(expiresAt.TTL() / tw.interval)
	slot := (tw.current + ticks) % tw.size

	tw.mu.Lock()
	tw.slots[slot] = append(tw.slots[slot], key)
	tw.mu.Unlock()

	return nil
}

// Get implements ExpirationManager.
func (tw *TimingWheelExpirationManager) Get(key Key) (Expiration, errors.Error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	// Check all slots for the key
	for i := 0; i < tw.size; i++ {
		for _, k := range tw.slots[i] {
			if k.String() == key.String() {
				// Return a dummy expiration, as we don't track individual expirations in this implementation
				return FromDuration(tw.interval), nil
			}
		}
	}
	return nil, NewStorageError("key-not-found", "Key not found", nil)
}

func (tw *TimingWheelExpirationManager) Cancel(key Key) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	// Remove the key from all slots
	for i := 0; i < tw.size; i++ {
		for j, k := range tw.slots[i] {
			if k.String() == key.String() {
				// Remove the key from the slot
				tw.slots[i] = append(tw.slots[i][:j], tw.slots[i][j+1:]...)
				break // Key found and removed, no need to check
			}
		}
	}
}

// run advances the wheel at each tick and deletes keys in the slot
func (tw *TimingWheelExpirationManager) Start(expire ExpireFunc) {
	ticker := time.NewTicker(tw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tw.mu.Lock()
			keys := tw.slots[tw.current]
			tw.slots[tw.current] = nil
			tw.mu.Unlock()

			// delete expired keys
			for _, key := range keys {
				go expire(key)
			}

			// advance wheel
			tw.current = (tw.current + 1) % tw.size

		case <-tw.stop:
			return
		}
	}
}

// Stop the timing wheel
func (tw *TimingWheelExpirationManager) Stop() {
	close(tw.stop)
}

var _ ExpirationManager = (*TimingWheelExpirationManager)(nil)

type PersistentExpirationManager struct {
	// The key space where keys are associated with their expiration times.
	space KeySpace[Expiration]

	// The underlying expiration manager (e.g. timing wheel or priority queue)
	inner ExpirationManager
}

// NewPersistentExpirationManager creates a new PersistentExpirationManager.
// It uses the provided KeySpace to manage key expirations and an inner ExpirationManager for scheduling expirations.
func NewPersistentExpirationManager(space KeySpace[Expiration], inner ExpirationManager) *PersistentExpirationManager {
	return &PersistentExpirationManager{
		space: space,
		inner: inner,
	}
}

// Schedule schedules a key to expire at the given time.
// It stores the expiration in the key space and schedules it with the inner manager.
func (m *PersistentExpirationManager) Schedule(key Key, expiresAt Expiration) errors.Error {
	err := m.inner.Schedule(key, expiresAt)
	if err != nil {
		return err
	}

	return m.space.Set(key, expiresAt)
}

// Get retrieves the expiration for a key from the key space.
func (m *PersistentExpirationManager) Get(key Key) (Expiration, errors.Error) {
	expiration, err := m.space.Get(key)
	if err != nil {
		if err.Code() == KeyNotFoundCode {
			return Never, nil // Key not found, return Never expiration
		}
		return nil, NewKeyError("get-expiration-failed", "Failed to get expiration for key", err).
			WithMetadata("key", key.String())
	}
	return expiration, nil
}

// Cancel removes a key's expiration.
// It cancels the expiration in the inner manager and removes the key from the key space.
func (m *PersistentExpirationManager) Cancel(key Key) {
	m.inner.Cancel(key)
	m.space.Del(key) // Error can be ignored, as it will not exist if it was never scheduled
}

// Start initializes the expiration manager by loading existing keys and their expirations from the key space.
// It schedules these keys with the inner expiration manager.
// The expire function will be called when a key expires, and it should handle the deletion logic
// (e.g., removing the key from storage).
// It panics if it fails to load keys or expirations from the key space.
func (m *PersistentExpirationManager) Start(expire ExpireFunc) {
	// Load all existing keys and their expirations from the key space
	keys, err := m.space.Keys("*")
	if err != nil {
		panic(NewKeyError("load-error", "Failed to load keys for expiration manager", err))
	}

	for _, key := range keys {
		expiration, err := m.space.Get(key)
		if err != nil {
			if err.Code() == KeyNotFoundCode {
				continue // Key not found, skip it
			}
			panic(NewKeyError("load-error", "Failed to get expiration for key", err).
				WithMetadata("key", key.String()))
		}

		m.inner.Schedule(key, expiration)
	}

	// Start the inner expiration manager
	m.inner.Start(func(key Key) {
		expire(key)
		m.space.Del(key)
	})
}

// Stop stops the expiration manager.
// It stops the inner expiration manager and cleans up any resources.
func (m *PersistentExpirationManager) Stop() {
	m.inner.Stop()
	// No need to clean up the key space, as it will be managed by the application lifecycle
}

var _ ExpirationManager = (*PersistentExpirationManager)(nil)

// HashExpirationKeySpace is a key space that manages expirations using a hash.
// It allows for efficient storage and retrieval of expiration times for multiple keys.
// It uses a single hash key to store all expirations, allowing for efficient lookups and updates.
type HashExpirationKeySpace struct {
	// mu protects the key space and cached expirations.
	mu sync.RWMutex

	// The key space where keys are associated with a hash of Expiration values.
	hks KeySpace[data.Hash[Expiration]]

	// The key under which the expiration hash is stored.
	// This allows for a single hash to manage expirations for multiple keys.
	key Key

	// expirations is a cached hash of expirations for the keys.
	// It is used to avoid frequent lookups in the key space.
	// It is updated when keys are added or removed.
	expirations data.Hash[Expiration] // Cached hash of expirations
}

func NewHashExpirationKeySpace(hks KeySpace[data.Hash[Expiration]], key Key) *HashExpirationKeySpace {
	return &HashExpirationKeySpace{
		hks:         hks,
		key:         key,
		expirations: data.NewDataHash(newExpirationCodec()),
	}
}

// Set stores a key with its expiration time.
// The expiration is stored in the hash and the hash is persisted to the key space.
func (hks *HashExpirationKeySpace) Set(key Key, value Expiration) errors.Error {
	hks.mu.Lock()
	defer hks.mu.Unlock()

	// Load the hash if not already loaded
	if err := hks.loadHashIfNeeded(); err != nil {
		return err
	}

	// Set the expiration in the hash
	hks.expirations.Set(key.String(), value)

	// Persist the hash to the key space
	return hks.hks.Set(hks.key, hks.expirations)
}

// Get retrieves the expiration for a given key.
func (hks *HashExpirationKeySpace) Get(key Key) (Expiration, errors.Error) {
	hks.mu.RLock()
	defer hks.mu.RUnlock()

	// Load the hash if not already loaded
	if err := hks.loadHashIfNeeded(); err != nil {
		return Never, err
	}

	// Get the expiration from the hash
	expiration := hks.expirations.Get(key.String())
	if expiration == Never {
		return Never, NewKeyNotFoundError(key, nil)
	}

	return expiration, nil
}

// Del removes the specified keys from the expiration hash.
func (hks *HashExpirationKeySpace) Del(keys ...Key) (int, errors.Error) {
	hks.mu.Lock()
	defer hks.mu.Unlock()

	// Load the hash if not already loaded
	if err := hks.loadHashIfNeeded(); err != nil {
		return 0, err
	}

	count := 0
	for _, key := range keys {
		// Check if the key exists first
		existing := hks.expirations.Get(key.String())
		if existing != Never {
			// Set it to Never to effectively "delete" it
			hks.expirations.Set(key.String(), Never)
			count++
		}
	}

	// Persist the hash to the key space if any keys were removed
	if count > 0 {
		if err := hks.hks.Set(hks.key, hks.expirations); err != nil {
			return 0, err
		}
	}

	return count, nil
}

// Keys retrieves all keys in the expiration hash that match the given pattern.
// Since the Hash interface doesn't provide a Keys method, we need to work around this limitation.
// This is a simplified implementation that doesn't support pattern matching.
func (hks *HashExpirationKeySpace) Keys(pattern string) ([]Key, errors.Error) {
	hks.mu.RLock()
	defer hks.mu.RUnlock()

	// Load the hash if not already loaded
	if err := hks.loadHashIfNeeded(); err != nil {
		return nil, err
	}

	// Since we can't iterate over the hash directly, we return an empty slice
	// In a real implementation, you'd need to extend the Hash interface to support iteration
	// or store the keys separately
	return []Key{}, nil
}

// Scan retrieves a limited number of keys matching a given glob pattern.
func (hks *HashExpirationKeySpace) Scan(pattern string, count int) ([]Key, errors.Error) {
	keys, err := hks.Keys(pattern)
	if err != nil {
		return nil, err
	}

	// Limit the results to the requested count
	if count > 0 && len(keys) > count {
		keys = keys[:count]
	}

	return keys, nil
}

// loadHashIfNeeded loads the hash from the key space if it hasn't been loaded yet.
// This method assumes the caller holds the appropriate lock.
func (hks *HashExpirationKeySpace) loadHashIfNeeded() errors.Error {
	// For now, always try to load the hash from the key space
	// In a real implementation, you'd want to track whether it's been loaded
	hash, err := hks.hks.Get(hks.key)
	if err != nil {
		// If the key doesn't exist, that's okay - we'll start with an empty hash
		if err.Code() == KeyNotFoundCode {
			return nil
		}
		return err
	}
	hks.expirations = hash
	return nil
}

var _ KeySpace[Expiration] = (*HashExpirationKeySpace)(nil)

func init() {
	gob.Register(expiration(time.Time{})) // Register the expiration type for gob encoding
}
