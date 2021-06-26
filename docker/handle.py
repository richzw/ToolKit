import sys, signal, os
from multiprocessing import Process

# Source: https://petermalmgren.com/pid-1-child-processes-docker/
from child import run

processes = []

def graceful_exit(_signum, _frame):
    print("Gracefully shutting down")
    for p in processes:
        print(f"Shutting down process {p.pid}")
        p.terminate()
    # we’ll make use of the multiprocessing.Process.join method, which calls os.waitpid() under the hood for UNIX systems.
    # We’ll give each process 60 seconds to finish up work, and then send it a SIGKILL.
    for p in processes:
        print(f"Waiting for process {p.pid}")
        p.join(60)
    # Having an indication of SIGKILL in our application logs is really handy for debugging misbehaving child processes.
	if p.exitcode is None:
            print(f"Sending SIGKILL to process {p.pid}")
            p.kill()
    sys.exit(0)

if __name__ == "__main__":
    for _ in range(10):
        proc = Process(target=run)
        proc.start()
        processes.append(proc)
    print(f"Parent process started: {os.getpid()}")
    signal.signal(signal.SIGTERM, graceful_exit)
    signal.signal(signal.SIGINT, graceful_exit)

    while True:
        # os.wait() and handle unexpected exits
        # Passing -1 to os.waitpid() tells it to wait on any process,
        # and passing in the option os.WNOHANG has it return immediately if no child process has exited.
        pid, status = os.waitpid(-1, os.WNOHANG)
        time.sleep(0.5)
	print("Still here!")