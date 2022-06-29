package errors

import "fmt"

func ExampleNew() {
	internal := fmt.Errorf("SSH connection failed due to unknown character in the signature")
	external := fmt.Errorf("invalid signature")
	// creates a new SSH errors from two errors, internal e an external.
	err := New(internal, external)

	fmt.Println(err)
	// Output: SSH connection failed due to unknown character in the signature
}

func ExampleGetInternal_from_error_type() {
	internal := fmt.Errorf("SSH connection failed due to unknown character in the signature")
	external := fmt.Errorf("invalid signature")
	// creates a new SSH errors from two errors, internal e an external.
	err := New(internal, external)

	fmt.Println(GetInternal(err))
	// Output: SSH connection failed due to unknown character in the signature
}

func ExampleGetInternal_no_error_type() {
	// creates a generic error.
	err := fmt.Errorf("SSH connection failed due to unknown character in the signature")

	fmt.Println(GetInternal(err))
	// Output: SSH connection failed due to unknown character in the signature
}

func ExampleGetExternal_from_error_type() {
	internal := fmt.Errorf("SSH connection failed due to unknown character in the signature")
	external := fmt.Errorf("invalid signature")
	// creates a new SSH errors from two errors, internal e an external.
	err := New(internal, external)

	fmt.Println(GetExternal(err))
	// Output: invalid signature
}

func ExampleGetExternal_no_error_type() {
	// creates a generic error.
	err := fmt.Errorf("SSH connection failed due to unknown character in the signature")

	fmt.Println(GetExternal(err))
	// Output: SSH connection failed due to unknown character in the signature
}

func ExampleIs_both_true() {
	internal := fmt.Errorf("SSH connection failed due to unknown character in the signature")
	external := fmt.Errorf("invalid signature")
	// creates a new SSH errors from two errors, internal e an external.
	err := New(internal, external)

	fmt.Println(Is(err, internal))
	fmt.Println(Is(err, external))
	// Output: true
	// true
}

func ExampleIs_true_false() {
	internal := fmt.Errorf("SSH connection failed due to unknown character in the signature")
	external := fmt.Errorf("invalid signature")
	// creates a new SSH errors from two errors, internal e an external.
	err := New(internal, fmt.Errorf("external generic error"))

	fmt.Println(Is(err, internal))
	fmt.Println(Is(err, external))
	// Output: true
	// false
}

func ExampleIs_false_true() {
	internal := fmt.Errorf("SSH connection failed due to unknown character in the signature")
	external := fmt.Errorf("invalid signature")
	// creates a new SSH errors from two errors, internal e an external.
	err := New(fmt.Errorf("internal generic error"), external)

	fmt.Println(Is(err, internal))
	fmt.Println(Is(err, external))
	// Output: false
	// true
}
