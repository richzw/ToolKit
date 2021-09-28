
- [interfaces are types](https://pauldigian.hashnode.dev/advanced-go-improve-your-code-using-interfaces-effectively)
  
  the first token when declaring a new interface is indeed type
  ```go
  type notImportant struct {}
  
  func (n *notImportant) Foo() bool {
      return true
  }
  
  func ValidFoo(f Fooer) bool {
      return f.Foo()
  }
  
  ValidFoo(&notImportat{})
  ```
  Note how in the body of the function we use only use methods defines in the interface and not other methods that the struct may implement.

- Interfaces are implicit

  concrete types do not have to declare if they implement a specific interface or not.

- Interfaces for condition dependent code

  
