PID 1 Responsibilities
--------

PID 1, or init process behavior in Linux.

PID 1 processes behave differently than other processes in a few different ways:

- They don’t get the default system signal handlers, so they are responsible for defining their own
- They are responsible for reaping zombie processes, or processes that have terminated but whose exit statuses haven’t been read
- Most importantly (for this scenario), they are responsible for adopting orphaned children and waiting on them to finish

PID 1 Responsibility Checklist
---------

When writing code that runs in a Docker container, there are four questions you’ll want to answer:

- Does your process respond to SIGTERM signals and gracefully shutdown?
- If spawning child processes, does your parent forward signals to them?
- If spawning child processes, does your parent process call os.wait() to see if they unexpectedly die?
- If spawning child processes, do you wait on all of your processes to exit before exiting yourself?
- Alternatively, if you don’t wait to be responsible for these things you can set the `ENTRYPOINT` of your Docker container to be a lightweight `init` process, such as `tini` which takes care of these for you.

What is tini
---------

- https://github.com/krallin/tini/issues/8
- https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/