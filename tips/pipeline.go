package tips

import (
	"errors"
	"context"
	"io"
	"log"
	"net/http"
	"os"
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