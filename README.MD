jz
==

JavaScript / TypeScript ES6 Virtual Machine in pure Go.
Based on: 
* https://github.com/dop251/goja
* https://github.com/dop251/goja_nodejs
* https://github.com/jvatic/goja-babel


### Example 
`cmd/example1/main.go`

```go
package main

import (
  "github.com/andrylavr/jz"
  "log"
)
import _ "embed"

//go:embed example1.js
var example1 string

func main() {
  vm := jz.New()
  vm.ImportMap["three"] = "https://threejs.org/build/three.js"
  vm.ImportMap["three/addons/"] = "https://threejs.org/examples/jsm/"

  _, err := vm.RunScript("example1.js", example1)
  if err != nil {
    log.Println(err)
  }
}

```

See more docs here: https://github.com/dop251/goja