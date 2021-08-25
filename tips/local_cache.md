
Local Cache with Sync.Once
------------------

`sync.Once` will wait until the execution inside the first `.Do` completes. This makes it incredibly useful when performing relatively expensive operations that you would typically cache in a map.

- naive cache

    ```go
    type QueryClient struct {
        cache map[string][]byte
        mutex *sync.Mutex
    }
    
    func (c *QueryClient) DoQuery(name string) []byte {
        // Check if the result is already cached.
        c.mutex.Lock()
        if cached, found := c.cache[name]; found {
            c.mutex.Unlock()
            return cached, nil
        }
        c.mutex.Unlock()
    
        // Make the request if it's uncached.
        resp, err := http.Get("https://upstream.api/?query=" + url.QueryEscape(name))
        // Error handling and resp.Body.Close omitted for brevity.
        result, err := ioutil.ReadAll(resp)
    
        // Store the result in the cache.
        c.mutex.Lock()
        c.cache[name] = result
        c.mutex.Unlock()
    
        return result
    }
    ```
  what happens if there are two calls to `DoQuery` that happen simultaneously? The calls would race, neither would see the cache is populated, and both would perform the HTTP request to upstream.api unnecessarily, when only one would need to complete it.
    
- Ugly but better solution?
    
    ```go
    type CacheEntry struct {
        data []byte
        wait <-chan struct{}
    }
    
    type QueryClient struct {
        cache map[string]*CacheEntry
        mutex *sync.Mutex
    }
    
    func (c *QueryClient) DoQuery(name string) []byte {
        // Check if the operation has already been started.
        c.mutex.Lock()
        if cached, found := c.cache[name]; found {
            c.mutex.Unlock()
            // Wait for it to complete.
            <-cached.wait
            return cached.data, nil
        }
    
        entry := &CacheEntry{
            data: result,
            wait: make(chan struct{}),
        }
        c.cache[name] = entry
        c.mutex.Unlock()
    
        // Make the request if it's uncached.
        resp, err := http.Get("https://upstream.api/?query=" + url.QueryEscape(name))
        // Error handling and resp.Body.Close omitted for brevity
        entry.data, err = ioutil.ReadAll(resp)
    
        // Signal that the operation is complete, receiving on closed channels
        // returns immediately.
        close(entry.wait)
    
        return entry.data
    }
    ```

- Apply sync.One
    
    ```go
    type CacheEntry struct {
        data []byte
        once *sync.Once
    }
    
    type QueryClient struct {
        cache map[string]*CacheEntry
        mutex *sync.Mutex
    }
    
    func (c *QueryClient) DoQuery(name string) []byte {
        c.mutex.Lock()
        entry, found := c.cache[name]
        if !found {
            // Create a new entry if one does not exist already.
            entry = &CacheEntry{
                once: new(sync.Once),
            }
            c.cache[name] = entry
        }
        c.mutex.Unlock()
    
        // Now when we invoke `.Do`, if there is an on-going simultaneous operation,
        // it will block until it has completed (and `entry.data` is populated).
        // Or if the operation has already completed once before,
        // this call is a no-op and doesn't block.
        entry.once.Do(func() {
            resp, err := http.Get("https://upstream.api/?query=" + url.QueryEscape(name))
            // Error handling and resp.Body.Close omitted for brevity
            entry.data, err = ioutil.ReadAll(resp)
        })
    
        return entry.data
    }
    ```
    
  `sync.One` with context

    ````go
    c := make(chan bool, 1)
    go func() {
        once.Do(f)
        c <- true
    }()
    select {
        case <-c:
        case <-ctxt.Done():
            return
    }
    ````

- TODO

  Another mechanism similar to `sync.Once` is `golang.org/x/sync/singleflight`. However singleflight only deduplicate requests that are in-flight (i.e. doesn't cache persistently). singleflight however may be cleaner to implement with `contexts` compared to `sync.Once` (through the use of a `select` and `ctx.Done()`), in production environments this may be important as to be able to cancel out with a context. The pattern with singleflight is quite similar to `sync.Once` but you would early return if a value is present inside the map.
