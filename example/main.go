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
		// During reads, a serializer converts the record back into the event.
		repository.WithSerializer(&gob.Serializer{}),

		// A store reads/writes records.
		repository.WithStore(memory.New()),

		// Observers are called when a command is successfully applied and
		// an event emitted. This would be a good hook to help subscribers
		// learn about state changes if you have subscriptions over websockets
		// for example. In this example, we'll simply log to demonstrate.
		repository.WithObservers(func(event eventsource.Event) {
			switch v := event.(type) {
			case *person.CreateEvent:
				// The event represents all of the data stored for and includes
				// more than what we log here.
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

	fmt.Printf("Request issues a command to Create Person...\n\n")

	rsp, err := personService.Create(ctx, person.CreateRequest{
		Name:  "Big Bird",
		Email: "b.bird@seasame-street.com",
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Read the resource back out left-folding over events to produce
	// our aggregate.
	fmt.Printf("Loading the resource and building the aggregate\n\n")

	aggregate, err := personService.Load(ctx, rsp.Person.ID)
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	fmt.Printf(`Loaded Person from Store
	Name: %q
	Email: %q
	ID: %q
	Version: %d
	Created By: %q
	Created By User ID: %q
	Created At: %q,
`,
		aggregate.Name,
		aggregate.Email,
		aggregate.ID,
		aggregate.Version,
		aggregate.CreateAudit.CreatedBy,
		aggregate.CreateAudit.CreatedByID,
		aggregate.CreateAudit.Created,
	)
}
