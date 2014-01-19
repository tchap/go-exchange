# go-exchange #

**Documentation**: [GoDoc](http://godoc.org/github.com/tchap/go-exchange/exchange)<br />
**Build Status**: [![Build Status](https://travis-ci.org/tchap/go-exchange.png?branch=master)](https://travis-ci.org/tchap/go-exchange)<br >
**Test Coverage**: TBD

## About ##

go-exchange is an in-process message (or event) exchange for Go (Golang), featuring:

* **Publish-Subscribe** with topic-based prefix subscriptions.

### State of the Project ###

This is a very young project and the API can and probably will change.

Any ideas on how to make the API or anything else better are welcome.

## Usage ##

```go
import "github.com/tchap/go-exchange/exchange"
```

### Patterns ###

This section describes the messaging patterns that go-exchange implements.

#### Publish-Subscribe ####

Publish-subscribe is based on topics and event handlers can be registered for
particular topic prefixes. What that means is that an event handler registered
for `git` topic will be as well activated when an events of kind `git.push` is
received.

##### Example #####

```go
ex := exchange.New()

ex.Subscribe(ex.Topic("git"), func(topic exchange.Topic, event exchange.Event) {
	fmt.Printf("Event received:\n  topic: %q\n  body:  %v\n", topic, event)
})  

ex.Publish(exchange.Topic("git.push"), "b839dc656e3e78647c09453b33652b389e37c07a")

ex.Terminate()
// Output:
// Event received:
//   topic: "git.push"
//   body:  b839dc656e3e78647c09453b33652b389e37c07a
```

## Ideas

* Limit how many handlers can be running at the same time.

## Related Projects ##

* [go-patricia](https://github.com/tchap/go-patricia) - trie that is being used
  in this project for managing subscriptions

## License ##

MIT, see the `LICENSE` file.

[![Gittip Badge](http://img.shields.io/gittip/alanhamlett.png)](https://www.gittip.com/tchap/ "Gittip Badge")

[![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/tchap/go-exchange/trend.png)](https://bitdeli.com/free "Bitdeli Badge")
