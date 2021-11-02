
- [Applying Modern Go Concurrency Patterns to Data Pipelines](https://medium.com/amboss/applying-modern-go-concurrency-patterns-to-data-pipelines-b3b5327908d4)
  - A Simple Pipeline
    ```go
    func producer(strings []string) (<-chan string, error) {
        outChannel := make(chan string)
    
        for _, s := range strings {
            outChannel <- s
        }
    
        return outChannel, nil
    }
    
    func sink(values <-chan string) {
        for value := range values {
            log.Println(value)
        }
    }
    
    func main() {
        source := []string{"foo", "bar", "bax"}
    
        outputChannel, err := producer(source)
        if err != nil {
            log.Fatal(err)
        }
    
        sink(outputChannel)
    }
    ```
    you run this with go run main.go you'll see a deadlock
    - The channel returned by producer is not buffered, meaning you can only send values to the channel if someone is receiving values on the other end. But since `sink` is called later in the program, there is no receiver at the point where `outChannel <- s` is called, causing the deadlock.
    - fix it
      - either making the channel buffered, in which case the deadlock will occur once the buffer is full
      - or by running the producer in a Go routine. 
      - whoever creates the channel is also in charge of closing it.
  - Graceful Shutdown With Context
    - with context
    ```go
    func producer(ctx context.Context, strings []string) (<-chan string, error) {
         outChannel := make(chan string)
     
         go func() {
             defer close(outChannel)
    
             for _, s := range strings {
                 time.Sleep(time.Second * 3)
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- s
                }
             }
         }()
     
         return outChannel, nil
     }
     
    
    func sink(ctx context.Context, values <-chan string) {
        for {
            select {
            case <-ctx.Done():
                log.Print(ctx.Err().Error())
                return
            case val, ok := <-values:
    +			log.Print(val)  // for debug
                if ok {
                    log.Println(val)
                }
            }
         }
     }
     
     func main() {
         source := []string{"foo", "bar", "bax"}
     
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()
    
        go func() {
            time.Sleep(time.Second * 5)
            cancel()
        }()
    
        outputChannel, err := producer(ctx, source)
         if err != nil {
             log.Fatal(err)
         }
     
    
        sink(ctx, outputChannel)
     }
    ```
    Issues:
    - This will flood our terminal with empty log messages, like this: 2021/09/08 12:29:30. Apparently the for loop in sink keeps running forever
    
    [Reason](https://golang.org/ref/spec#Receive_operator)
    - A receive operation on a closed channel can always proceed immediately, yielding the element type’s zero value after any previously sent values have been received.

    Fix it
    - The value of ok is true if the value received was delivered by a successful send operation to the channel, or false if it is a zero value generated because the channel is closed and empty.
     ```go
     func sink(ctx context.Context, values <-chan string) {
     
              case val, ok := <-values:
                  if ok {
                      log.Println(val)
                 } else {
                     return
                  }
              }
          }
     ```
  - Adding Parallelism with Fan-Out and Fan-In
    - sending values to a closed channel is a panic
    ```go
    func producer(ctx context.Context, strings []string) (<-chan string, error) {
        outChannel := make(chan string)
    
        go func() {
            defer close(outChannel)
    
            for _, s := range strings {
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- s
                }
            }
        }()
    
        return outChannel, nil
    }
    
    func transformToLower(ctx context.Context, values <-chan string) (<-chan string, error) {
        outChannel := make(chan string)
    
        go func() {
            defer close(outChannel)
    
            for s := range values {
                time.Sleep(time.Second * 3)
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- strings.ToLower(s)
                }
            }
        }()
    
        return outChannel, nil
    }
    
    func transformToTitle(ctx context.Context, values <-chan string) (<-chan string, error) {
        outChannel := make(chan string)
    
        go func() {
            defer close(outChannel)
    
            for s := range values {
                time.Sleep(time.Second * 3)
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- strings.ToTitle(s)
                }
            }
        }()
    
        return outChannel, nil
    }
    
    func sink(ctx context.Context, values <-chan string) {
        for {
            select {
            case <-ctx.Done():
                log.Print(ctx.Err().Error())
                return
            case val, ok := <-values:
                if ok {
                    log.Println(val)
                } else {
                    return
                }
            }
        }
    }
    
    func main() {
        source := []string{"FOO", "BAR", "BAX"}
    
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()
    
        outputChannel, err := producer(ctx, source)
        if err != nil {
            log.Fatal(err)
        }
    
        stage1Channels := []<-chan string{}
    
        for i := 0; i < runtime.NumCPU(); i++ {
            lowerCaseChannel, err := transformToLower(ctx, outputChannel)
            if err != nil {
                log.Fatal(err)
            }
            stage1Channels = append(stage1Channels, lowerCaseChannel)
        }
    
        stage1Merged := mergeStringChans(ctx, stage1Channels...)
        stage2Channels := []<-chan string{}
    
        for i := 0; i < runtime.NumCPU(); i++ {
            titleCaseChannel, err := transformToTitle(ctx, stage1Merged)
            if err != nil {
                log.Fatal(err)
            }
            stage2Channels = append(stage2Channels, titleCaseChannel)
        }
    
        stage2Merged := mergeStringChans(ctx, stage2Channels...)
        sink(ctx, stage2Merged)
    }
    
    func mergeStringChans(ctx context.Context, cs ...<-chan string) <-chan string {
        var wg sync.WaitGroup
        out := make(chan string)
    
        output := func(c <-chan string) {
            defer wg.Done()
            for n := range c {
                select {
                case out <- n:
                case <-ctx.Done():
                    return
                }
            }
        }
    
        wg.Add(len(cs))
        for _, c := range cs {
            go output(c)
        }
    
        go func() {
            wg.Wait()
            close(out)
        }()
    
        return out
    }
    ```
  - Error Handling
    - The most common way of propagating errors that I’ve seen is through a separate error channel. Unlike the value channels that connect pipeline stages, the error channels are not passed to downstream stages.
  - Removing Boilerplate With Generics
  - Maximum Efficiency With Semaphores
    - What if our input list only had a single element in it? Then we only need a single Go routine, not NumCPU() Go routines. 
    - Instead of creating a fixed number of Go routines, we will range over the input channel. For every value we receive from it, we will spawn a Go routine (see the example in the semaphore package)
    
    ```go
    package tips
    
    import (
        "errors"
        "context"
        "log"
        "runtime"
        "strings"
        "time"
    
        "golang.org/x/sync/semaphore"
    )
    
    func producer(ctx context.Context, strings []string) (<-chan string, error) {
        outChannel := make(chan string)
    
        go func() {
            defer close(outChannel)
    
            for _, s := range strings {
                select {
                case <-ctx.Done():
                    return
                default:
                    outChannel <- s
                }
            }
        }()
    
        return outChannel, nil
    }
    
    func sink(ctx context.Context, cancelFunc context.CancelFunc, values <-chan string, errors <-chan error) {
        for {
            select {
            case <-ctx.Done():
                log.Print(ctx.Err().Error())
                return
            case err := <-errors:
                if err != nil {
                    log.Println("error: ", err.Error())
                    cancelFunc()
                }
            case val, ok := <-values:
                if ok {
                    log.Printf("sink: %s", val)
                } else {
                    log.Print("done")
                    return
                }
            }
        }
    }
    
    func step[In any, Out any](
        ctx context.Context,
        inputChannel <-chan In,
        outputChannel chan Out,
        errorChannel chan error,
        fn func(In) (Out, error),
    ) {
        defer close(outputChannel)
    
        limit := runtime.NumCPU()
        sem1 := semaphore.NewWeighted(limit)
    
        for s := range inputChannel {
            select {
            case <-ctx.Done():
                log.Print("1 abort")
                break
            default:
            }
    
            if err := sem1.Acquire(ctx, 1); err != nil {
                log.Printf("Failed to acquire semaphore: %v", err)
                break
            }
    
            go func(s In) {
                defer sem1.Release(1)
                time.Sleep(time.Second * 3)
    
                result, err := fn(s)
                if err != nil {
                    errorChannel <- err
                } else {
                    outputChannel <- result
                }
            }(s)
        }
    
        if err := sem1.Acquire(ctx, limit); err != nil {
            log.Printf("Failed to acquire semaphore: %v", err)
        }
    }
    
    func main() {
        source := []string{"FOO", "BAR", "BAX"}
    
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()
    
        readStream, err := producer(ctx, source)
        if err != nil {
            log.Fatal(err)
        }
    
        stage1 := make(chan string)
        errorChannel := make(chan error)
    
        transformA := func(s string) (string, error) {
            return strings.ToLower(s), nil
        }
    
        go func() {
            step(ctx, readStream, stage1, errorChannel, transformA)
        }()
    
        stage2 := make(chan string)
    
        transformB := func(s string) (string, error) {
            if s == "foo" {
                return "", errors.New("oh no")
            }
    
            return strings.Title(s), nil
        }
    
        go func() {
            step(ctx, stage1, stage2, errorChannel, transformB)
        }()
    
        sink(ctx, cancel, stage2, errorChannel)
    }
    ```

- [Handling 1 Million Requests per Minute with Go](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/)
  - The web handler would receive a JSON document that may contain a collection of many payloads that needed to be written to Amazon S3
  - Naive approach to GO routine
    ```go
    func payloadHandler(w http.ResponseWriter, r *http.Request) {
    
        if r.Method != "POST" {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }
    
        // Read the body into a string for json decoding
        var content = &PayloadCollection{}
        err := json.NewDecoder(io.LimitReader(r.Body, MaxLength)).Decode(&content)
        if err != nil {
            w.Header().Set("Content-Type", "application/json; charset=UTF-8")
            w.WriteHeader(http.StatusBadRequest)
            return
        }
    
        // Go through each payload and queue items individually to be posted to S3
        for _, payload := range content.Payloads {
            go payload.UploadToS3()   // <----- DON'T DO THIS
        }
    
        w.WriteHeader(http.StatusOK)
    }
    ```
  - Try again
    - create a buffered channel where we could queue up some jobs and upload them to S3
     ```go
     var Queue chan Payload
     
     func init() {
         Queue = make(chan Payload, MAX_QUEUE)
     }
     
     func payloadHandler(w http.ResponseWriter, r *http.Request) {
         ...
         // Go through each payload and queue items individually to be posted to S3
         for _, payload := range content.Payloads {
             Queue <- payload
         }
         ...
     }
     func StartProcessor() {
         for {
             select {
             case job := <-Queue:
                 job.payload.UploadToS3()  // <-- STILL NOT GOOD
             }
         }
     }
     ```
    Issue: since the rate of incoming requests were much larger than the ability of the single processor to upload to S3, our buffered channel was quickly reaching its limit and blocking the request handler ability to queue more items.
  - Better solution
    - create a 2-tier channel system, one for queuing jobs and another to control how many workers operate on the JobQueue concurrently.

    ```go
    var (
    	MaxWorker = os.Getenv("MAX_WORKERS")
    	MaxQueue  = os.Getenv("MAX_QUEUE")
    )
    
    // Job represents the job to be run
    type Job struct {
    	Payload Payload
    }
    
    // A buffered channel that we can send work requests on.
    var JobQueue chan Job
    
    // Worker represents the worker that executes the job
    type Worker struct {
    	WorkerPool  chan chan Job
    	JobChannel  chan Job
    	quit    	chan bool
    }
    
    func NewWorker(workerPool chan chan Job) Worker {
    	return Worker{
    		WorkerPool: workerPool,
    		JobChannel: make(chan Job),
    		quit:       make(chan bool)}
    }
    
    // Start method starts the run loop for the worker, listening for a quit channel in
    // case we need to stop it
    func (w Worker) Start() {
    	go func() {
    		for {
    			// register the current worker into the worker queue.
    			w.WorkerPool <- w.JobChannel
    
    			select {
    			case job := <-w.JobChannel:
    				// we have received a work request.
    				if err := job.Payload.UploadToS3(); err != nil {
    					log.Errorf("Error uploading to S3: %s", err.Error())
    				}
    
    			case <-w.quit:
    				// we have received a signal to stop
    				return
    			}
    		}
    	}()
    }
    
    // Stop signals the worker to stop listening for work requests.
    func (w Worker) Stop() {
    	go func() {
    		w.quit <- true
    	}()
    }
    
    // handler
    func payloadHandler(w http.ResponseWriter, r *http.Request) {
    
    	if r.Method != "POST" {
    		w.WriteHeader(http.StatusMethodNotAllowed)
    		return
    	}
    
    	// Read the body into a string for json decoding
    	var content = &PayloadCollection{}
    	err := json.NewDecoder(io.LimitReader(r.Body, MaxLength)).Decode(&content)
    	if err != nil {
    		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    		w.WriteHeader(http.StatusBadRequest)
    		return
    	}
    
    	// Go through each payload and queue items individually to be posted to S3
    	for _, payload := range content.Payloads {
    
    		// let's create a job with the payload
    		work := Job{Payload: payload}
    
    		// Push the work onto the queue.
    		JobQueue <- work
    	}
    
    	w.WriteHeader(http.StatusOK)
    }
    
    dispatcher := NewDispatcher(MaxWorker)
    dispatcher.Run()
    
    type Dispatcher struct {
    	// A pool of workers channels that are registered with the dispatcher
    	WorkerPool chan chan Job
    }
    
    func NewDispatcher(maxWorkers int) *Dispatcher {
    	pool := make(chan chan Job, maxWorkers)
    	return &Dispatcher{WorkerPool: pool}
    }
    
    func (d *Dispatcher) Run() {
    	// starting n number of workers
    	for i := 0; i < d.maxWorkers; i++ {
    		worker := NewWorker(d.pool)
    		worker.Start()
    	}
    
    	go d.dispatch()
    }
    
    func (d *Dispatcher) dispatch() {
    	for {
    		select {
    		case job := <-JobQueue:
    			// a job request has been received
    			go func(job Job) {
    				// try to obtain a worker job channel that is available.
    				// this will block until a worker is idle
    				jobChannel := <-d.WorkerPool
    
    				// dispatch the job to the worker job channel
    				jobChannel <- job
    			}(job)
    		}
    	}
    }
    ```









