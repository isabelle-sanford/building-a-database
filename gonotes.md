## FROM GTOWELL 245

toString equivalent on a struct Strc

```go
// this behaves much the same as the toString method in Java
func (item Strc) String() string {
	return fmt.Sprintf("Strcc A:%v  b:%v  c:%v", item.A, item.b, item.c);
}
```

// Arrays are passed by value. This means that the
// array and everything in the array is new pass
// Arrays also returned by value

// Not strictly necessary, but defining types for funtions being passed, or
// returned make the code more readable, IMHO
type af func(a int) int

```go
// summTR2 does tail recursion, but uses a private recursive function to do so!
// This gets past Go not alloing overloaded function names or default params!!
func summTR2(mx int) int {
	var ttr func(int, int) int // need to define the variable first so it is available for the recursion
	ttr = func(mm int, tot int) int {
		if (mm<=0) { return tot }
		return ttr(mm-1, tot+mm)
	}
	return ttr(mx, 0)
}
```

```go
// summTR2 does tail recursion, but uses a private recursive function to do so!
// This gets past Go not alloing overloaded function names or default params!!
func summTR2(mx int) int {
	var ttr func(int, int) int // need to define the variable first so it is available for the recursion
	ttr = func(mm int, tot int) int {
		if (mm<=0) { return tot }
		return ttr(mm-1, tot+mm)
	}
	return ttr(mx, 0)
}
```
