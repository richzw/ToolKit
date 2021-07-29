avoid race condition

- ticket storage

  something like atomic on linked list

```go
  type TicketStore struct {
      ticket *uint64
      done   *uint64
      slots  []string
  }

  func (ts *TicketStore) Put(s string) {
      t = atomic.AddUnit64(ts.ticket, 1) - 1
      slots[t] = s
      for !atomic.CompareAndSwapUint64(ts.done, t, t+1) {
        runtime.GoSched()
      }
  }

  func (ts *TicketStore) GetDone() []string {
    return ts.slots[:atomic.LoadUint64(ts.done) + 1]
  }
```
