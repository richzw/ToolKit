
- [Thread vs Process in Python](https://superfastpython.com/thread-vs-process/)
  - For CPU bound work, multiprocessing is always faster, presumably due to the GIL
    - Calculating points in a fractal.
    - Estimating Pi
    - Factoring primes.
    - Parsing HTML, JSON, etc. documents.
    - Processing text.
    - Running simulations.
  - Use Threads for for IO-Bound
    - Reading or writing a file from the hard drive.
    - Reading or writing to standard output, input, or error (stdin, stdout, stderr).
    - Printing a document.
    - Downloading or uploading a file.
    - Querying a server.
    - Querying a database.
    - Taking a photo or recording a video






