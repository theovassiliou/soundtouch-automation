package main

import "fmt"

func ExampleFormatFullVersion() {
	fmt.Println(FormatFullVersion("myApp", "1.0", "main", "b2fecc", "2023-01-02T14:22:23Z"))
	fmt.Println(FormatFullVersion("myApp", "1.0", "", "b2fecc", "2023-01-02T14:22:23Z"))
	fmt.Println(FormatFullVersion("myApp", "1.0", "", "", "2023-01-02T14:22:23Z"))
	fmt.Println(FormatFullVersion("myApp", "1.0", "", "", ""))

	// Output:
	// myApp 1.0 (git: main b2fecc) (build: 2023-01-02T14:22:23Z)
	// myApp 1.0 (git: unknown b2fecc) (build: 2023-01-02T14:22:23Z)
	// myApp 1.0 (build: 2023-01-02T14:22:23Z)
	// myApp 1.0

}
