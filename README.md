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

Run the server with `arbor`. It listens on port 7777 by default, but this can be changed by running with the port number as the first argument: `arbor 8080`.

#### Flags
- ruser         Running the server with the flag `-ruser foo` changes the username of the root user. Currently this will only change the root message spawned upon startup. The default username is "root".
- rid           The `-rid XXXX` flag changes the UUID of the server's root message. If unused the server will assign a random UUID to the the root message.
- rcontent      `-rcontent bar` changes the content of the root message. If left out the root message will say "Welcome to our server!".
- recent-size   The size of the recent list determines the number of new messages the server has cached to send to new clients upon connection. Starting the server with `-recent-size XXX` changes the starting number of the recents list. This flag **could** effect the optimization of the server. The lists defaults to a capacity of 100 items.
