
- [undefined timezone behavior](https://www.dolthub.com/blog/2021-09-03-golang-time-bugs/)

  There are some actual foot-guns hidden in the runtime:
  - Reads from a _nil_ map are fine, writes panic
  - Reads from a closed channel block, sends on a closed channel panic
  - nil sometimes does not _== nil_
  - Any panic in any goroutine kills the entire process
  - Can't safely store references to loop variables

  - Issue: `time.Date(1970, 1, 1, 0, 0, 0, 0, &time.Location{})`
  
    _time_ struct returned by the query had a different timezone than the expected value. Specifically, we weren't giving an initialized time.Location for the expected time object

    golang doesn't try to determine what your timezone info is without asking you. That would be rude! Instead, it waits for you to tell it to load this info explicitly, after which further calls to fetch the Local location will include this cached, detailed information. But calling Format on a time object, for some format strings (such as _time.RFC3339_), needs to load this information, so will do so implicitly as a side effect. After such a call to _time.Format_, all _time.Time_ structs with the time.Local location will include this info, when they didn't before.
  - Fix it
    ```go
    // We are doing structural equality tests on time.Time values in some of these
    // tests. The SQL value layer works with times in time.Local location, but
    // go standard library will return different values (which will have the same
    // behavior) depending on whether detailed timezone information has been loaded
    // for time.Local already. Here, we always load the detailed information so
    // that our structural equality tests will be reliable.
    var loadedLocalLocation *time.Location
    
    func LoadedLocalLocation() *time.Location {
        var err error
        loadedLocalLocation, err = time.LoadLocation(time.Local.String())
        if err != nil {
            panic(err)
        }
        if loadedLocalLocation == nil {
            panic("nil LoadedLocalLocation " + time.Local.String())
        }
        return loadedLocalLocation
    }
    
    time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).In(LoadedLocalLocation())
    ```

- [Go is pass-by-value](https://neilalexander.dev/2021/08/29/go-pass-by-value.html)
  - using _*type_ instead of _type_ and then taking the reference of a variable using _&variable_. In this case, passing a pointer into a function is still passing by value in the strictest sense, but it’s actually the pointer’s value itself that is being copied, not the thing that the pointer refers to.
  - Go has two classes of datatype: those with “value” semantics and those with “reference” semantics
    - “value” types include the usual int, string, byte rune, bool types 
    - “reference” types. However, maps, slices and channels
  - Go is still really passing-by-value in the truest sense of the term even with more complex types, it’s just that the references are being passed as values instead of the contents. However, the actual resulting behaviour can feel unintuitive at first and can result in difficult-to-trace errors and bugs within a program. Even built-in functions like append have the potential to lead you astray if you make incorrect assumptions about whether or not a copy will take place.





