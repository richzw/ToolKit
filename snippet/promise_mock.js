// 1. simple version  // https://medium.com/swlh/implement-a-simple-promise-in-javascript-20c9705f197a
class Promise {
    constructor(handler) {
        this.status = "pending";
        this.onFulfilledCallbacks = [];
        this.onRejectedCallbacks = [];

        const resolve = value => {
            if (this.status === "pending") {
                this.status = "fulfilled";
                this.value = value;
                this.onFulfilledCallbacks.forEach(fn => fn(value));
            }
        };

        const reject = value => {
            if (this.status === "pending") {
                this.status = "rejected";
                this.value = value;
                this.onRejectedCallbacks.forEach(fn => fn(value));
            }
        };

        try {
            handler(resolve, reject);
        } catch (err) {
            reject(err);
        }
    }

    then(onFulfilled, onRejected) {
        return new Promise((resolve, reject) => {
            if (this.status === "pending") {
                this.onFulfilledCallbacks.push(() => {
                    try {
                        const fulfilledFromLastPromise = onFulfilled(this.value);
                        if (fulfilledFromLastPromise instanceof Promise) {
                            fulfilledFromLastPromise.then(resolve, reject);
                        } else {
                            resolve(fulfilledFromLastPromise);
                        }
                    } catch (err) {
                        reject(err);
                    }
                });
                this.onRejectedCallbacks.push(() => {
                    try {
                        const rejectedFromLastPromise = onRejected(this.value);
                        if (rejectedFromLastPromise instanceof Promise) {
                            rejectedFromLastPromise.then(resolve, reject);
                        } else {
                            reject(rejectedFromLastPromise);
                        }
                    } catch (err) {
                        reject(err);
                    }
                });
            }

            if (this.status === "fulfilled") {
                try {
                    const fulfilledFromLastPromise = onFulfilled(this.value);
                    if (fulfilledFromLastPromise instanceof Promise) {
                        fulfilledFromLastPromise.then(resolve, reject);
                    } else {
                        resolve(fulfilledFromLastPromise);
                    }
                } catch (err) {
                    reject(err);
                }

            }

            if (this.status === "rejected") {
                try {
                    const rejectedFromLastPromise = onRejected(this.value);
                    if (rejectedFromLastPromise instanceof Promise) {
                        rejectedFromLastPromise.then(resolve, reject);
                    } else {
                        reject(rejectedFromLastPromise);
                    }
                } catch (err) {
                    reject(err);
                }
            }
        });

    }
}

// testing code
let p1 = new Promise((resolve, reject) => {
    setTimeout(() => resolve('resolved first one'), 1000);
});
p1.then((res) => {
    console.log(res);
    return new Promise(resolve => {
        setTimeout(() => resolve('resolved second one'), 1000);
    });
}).then(res => {
    console.log(res);
});

// 2. complex version
var PENDING = 0;
var FULFILLED = 1;
var REJECTED = 2;

function PromiseComplex(fn) {
  // store state which can be PENDING, FULFILLED or REJECTED
  var state = PENDING;

  // store value once FULFILLED or REJECTED
  var value = null;

  // store sucess & failure handlers
  var handlers = [];

  function fulfill(result) {
    state = FULFILLED;
    value = result;
    handlers.forEach(handle);
    handlers = null;
  }

  function reject(error) {
    state = REJECTED;
    value = error;
    handlers.forEach(handle);
    handlers = null;
  }

  function resolve(result) {
    try {
      var then = getThen(result);
      if (then) {
        doResolve(then.bind(result), resolve, reject)
        return
      }
      fulfill(result);
    } catch (e) {
      reject(e);
    }
  }

  function handle(handler) {
    if (state === PENDING) {
      handlers.push(handler);
    } else {
      if (state === FULFILLED &&
        typeof handler.onFulfilled === 'function') {
        handler.onFulfilled(value);
      }
      if (state === REJECTED &&
        typeof handler.onRejected === 'function') {
        handler.onRejected(value);
      }
    }
  }

  this.done = function (onFulfilled, onRejected) {
    // ensure we are always asynchronous
    setTimeout(function () {
      handle({
        onFulfilled: onFulfilled,
        onRejected: onRejected
      });
    }, 0);
  }

  doResolve(fn, resolve, reject);

  this.then = function (onFulfilled, onRejected) {
    var self = this;
    return new Promise(function (resolve, reject) {
      return self.done(function (result) {
        if (typeof onFulfilled === 'function') {
          try {
            return resolve(onFulfilled(result));
          } catch (ex) {
            return reject(ex);
          }
        } else {
          return resolve(result);
        }
      }, function (error) {
        if (typeof onRejected === 'function') {
          try {
            return resolve(onRejected(error));
          } catch (ex) {
            return reject(ex);
          }
        } else {
          return reject(error);
        }
      });
    });
  }
}

/**
 * Check if a value is a Promise and, if it is,
 * return the `then` method of that promise.
 *
 * @param {Promise|Any} value
 * @return {Function|Null}
 */
function getThen(value) {
  var t = typeof value;
  if (value && (t === 'object' || t === 'function')) {
    var then = value.then;
    if (typeof then === 'function') {
      return then;
    }
  }
  return null;
}

/**
 * Take a potentially misbehaving resolver function and make sure
 * onFulfilled and onRejected are only called once.
 *
 * Makes no guarantees about asynchrony.
 *
 * @param {Function} fn A resolver function that may not be trusted
 * @param {Function} onFulfilled
 * @param {Function} onRejected
 */
function doResolve(fn, onFulfilled, onRejected) {
  var done = false;
  try {
    fn(function (value) {
      if (done) return
      done = true
      onFulfilled(value)
    }, function (reason) {
      if (done) return
      done = true
      onRejected(reason)
    })
  } catch (ex) {
    if (done) return
    done = true
    onRejected(ex)
  }
}

