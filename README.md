# Arbor

Arbor is an experimental chat protocol that models a conversation
as a tree of messages instead of an ordered list. This means that
the conversation can organically diverge into several conversations
without the messages appearing interleaved.

Arbor is unbelievably primitive right now. With time, it may develop
into something usable, but be warned that it is not currently a delightful
user experience.

## Testing Arbor
If you'd like to see where things stand, you should be able to do the following:

```
go get github.com/arborchat/server/cmd/...
```

Run the server with `arbor`, it listens on port 7777 by default.
