package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aarongreenlee/eventsource"
	"github.com/aarongreenlee/eventsource/example/person"
	"github.com/aarongreenlee/eventsource/example/session"

	"github.com/aarongreenlee/eventsource/repository"
	"github.com/aarongreenlee/eventsource/serializer/gob"
	"github.com/aarongreenlee/eventsource/store/memory"
)

func main() {

	// Build our Person service and configure the underlying repository
	// as you like. Here, we're using a Memory store and a GOB serializer.
	//
	// An eventsource.Repository establishes both the Gob Serializer and
	// Memory Store as defaults but we configure them here to demonstrate
	// how you might configure your own service as you initialize.
	personService, err := person.NewService(
		// During writes, a serializer converts events to records.
		// During reads, a serialized converts the record back into the event.
		repository.WithSerializer(&gob.Serializer{}),
		// A store reads/writes records.
		repository.WithStore(memory.New()),
		// Observers are called when a command is successfully applied and
		// an event emitted. This would be a good hook to help subscribers
		// learn about state changes if you subscriptions over websockets
		// for example. In this example, we'll simply log.
		repository.WithObservers(func(event eventsource.Event) {
			switch v := event.(type) {
			case *person.CreateEvent:
				fmt.Printf("Event %q Observed\n\tName %q\n\tEmail: %q\n\n", event.EventType(), v.Name, v.Email)
			}
		}),
	)
	if err != nil {
		fmt.Printf("error building user service: %s\n", err)
		os.Exit(1)
	}

	// With our domain model service configured we can simulate some
	// work flow!
	ctx := session.Stub(context.Background()) // simulate a session

	fmt.Printf("Issuing command to Create Person...\n\n")
	rsp, err := personService.Create(ctx, person.CreateRequest{
		Name:  "Big Bird",
		Email: "b.bird@seasame-street.com",
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf(
		"Response\n\tName: %q\n\tID: %q\n\tVersion: %d\n\n",
		rsp.Person.Name,
		rsp.Person.ID,
		rsp.Person.Version,
	)
}
