Q: Why does 6.824 use Go for the labs?

A: Until a few years ago 6.824 used C++, which worked well. Go works a
little better for 6.824 labs for a couple reasons. Go is garbage
collected and type-safe, which eliminates some common classes of bugs.
Go has good support for threads (goroutines), and a nice RPC package,
which are directly useful in 6.824. Threads and garbage collection
work particularly well together, since garbage collection can
eliminate programmer effort to decide when the last thread using an
object has stopped using it. There are other languages with these
features that would probably work fine for 6.824 labs, such as Java.

Q: How do Go channels work? How does Go make sure they are
synchronized between the many possible goroutines?

A: You can see the source at https://golang.org/src/runtime/chan.go,
though it is not easy to follow.

At a high level, a chan is a struct holding a buffer and a lock.
Sending on a channel involves acquiring the lock, waiting (perhaps
releasing the CPU) until some thread is receiving, and handing off the
message. Receiving involves acquiring the lock and waiting for a
sender. You could implement your own channels with Go sync.Mutex and
sync.Cond.

Q: Do goroutines run in parallel? Can you use them to increase
performance?

A: Go's goroutines are the same as threads in other languages. The Go
runtime executes goroutines on all available cores, in parallel.
If only one core is available, the runtime will pre-emptively
time-share the core among goroutines.

You can use goroutines in a coroutine style, but they aren't
restricted to that style.

Q: What's the best way to have a channel that periodically checks for
input, since we don't want it to block but we also don't want to be
checking it constantly?

A: Try creating a separate goroutine for each channel, and have each
goroutine block on its channel. That's not always possible, but when
it works I find it's often the simplest approach.

Use select with a default case to check for channel input without
blocking. If you want to block for a short time, use select and
time.After(). Here are examples:

https://tour.golang.org/concurrency/6
https://gobyexample.com/timeouts

Q: Is Go used in industry?

A: You can see an estimate of how much different programming languages
are used here:

https://www.tiobe.com/tiobe-index/

Q: When should we use sync.WaitGroup instead of channels? and vice versa?

A: WaitGroup is fairly special-purpose; it's only useful when waiting
for a bunch of activities to complete. Channels are more
general-purpose; for example, you can communicate values over
channels. You can wait for multiple goroutines using channels, though it
takes a few more lines of code than with WaitGroup.

Q: What are some important/useful Go-specific concurrency patterns to know?

A: Here's a slide deck on this topic, from a Go expert:

https://talks.golang.org/2012/concurrency.slide

Q: How are slices implemented?

A: A slice is an object that contains a pointer to an array and a start and
end index into that array. This arrangement allows multiple slices to
share an underlying array, with each slice perhaps exposing a different
range of array elements.

Here's a more extended discussion:

  https://blog.golang.org/go-slices-usage-and-internals

I use slices a lot, but I rarely use them to share an array. I hardly
ever directly use arrays. I basically use slices as if they were arrays.
A Go slice is more flexible than a Go array since an array's size is
part of its type, whereas a function that takes a slice as argument can
take a slice of any length.


Q: How do we know when the overhead of spawning goroutines exceeds
the concurrency we gain from them?

A: It depends! If your machine has 16 cores, and you are looking for
CPU parallelism, you should have roughly 16 executable goroutines. If it
takes 0.1 second of real time to fetch a web page, and your network is
capable of transmitting 100 web pages per second, you probably need
about 10 goroutines concurrently fetching in order to use all of the
network capacity. Experimentally, as you increase the number of
goroutines, for a while you'll see increased throughput, and then you'll
stop getting more throughput; at that point you have enough goroutines
from the performance point of view. 

Q: How would one create a Go channel that connects over the Internet?
How would one specify the protocol to use to send messages?

A: A Go channel only works within a single program; channels cannot be
used to talk to other programs or other computers.

Have a look at Go's RPC package, which lets you talk to other Go
programs over the Internet:

  https://golang.org/pkg/net/rpc/

Q: For what types of applications would you recommend using Go?

A: One answer is that if you were thinking of using C or C++ or Java,
you could also consider using Go.

Another answer is that you should use the language that you know best
and are most productive in.

Another answer is that there are some specific problem domains, such as
big numeric computations, for which well-tailored languages exist (e.g.
Fortran or Python+numpy). For those situations, you should consider
using the corresponding specialized language. Otherwise use a
general-purpose language in which you feel comfortable, for example Go.

Another answer is that for some problem domains, it's important to use a
language that has extensive domain-specific libraries. Writing web
servers, for example, works best if you have easy access to libraries
that can parse HTTP requests &c.

Q: What are common debugging tools people use for go?

Q: fmt.Printf()

As far as I know there's not a great debugger for Go, though gdb can be
made to work:

https://golang.org/doc/gdb

In any case, I find fmt.Printf() much more generally useful than any
debugger I've ever used, regardless of language.

Q: When is it right to use a synchronous RPC call and when is it right to
use an asynchronous RPC call?

A: Most code needs the RPC reply before it can proceed; in that case it
makes sense to use synchronous RPC.

But sometimes a client wants to launch many concurrent RPCs; in that
case async may be better. Or the client wants to do other work while it
waits for the RPC to complete, perhaps because the server is far away
(so speed-of-light time is high) or because the server might not be
reachable so that the RPC suffers a long timeout period.

I have never used async RPC in Go. When I want to send an RPC but not
have to wait for the result, I create a goroutine, and have the
goroutine make a synchronous Call().
