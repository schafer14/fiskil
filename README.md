# What's here?

In a short time frame I had to be judicious about the scope of this project. I have decided to focus on the problem and leave the peripherals aside. Focusing on the core of the problem allows me to be specific in the model that I create, and ensure that there are no extraneous types that get mixed into the domain model.  

The collector has been included, while adapters for the broker and storage layers have been omitted. 

Testing:

```
go test ./...
```

# Worries

The collector is not taking full advantages of Go's concurrency priviledges. There are a number of ways to address this, but it's probably best to gets some concrete measurements of max system load and setup performance testing instead of blindly optimize.

Another concern is \*collector probably needs to be provided an error channel. The logic for how to handle exceptions (for instance if the ticker channel closes), should probably be delegated to a supervisor who can observe the whole system and make good decesions and potentially escalate to a human if need be.

These issues are why it's so valuable to be able to run acceptance tests in a production live environment. Answer questions of optimization and system resource requires concrete examples that would best be collected from running real load through the system. 

# Change in spec

There's a slight change in the specification in terms of frequency variable. I changed it from a duration to a channel because I believe that it should not be configurable and rather the storage layer should be able to control the frequency for flushes. This creates back pressuer and allows the collector to make decisions about how to process flush events when the storage layer is under stress (instead of just killing the DB...). This provides two benefits: fine grain handling of system resource depletion, and containment of system resource failure.

For this reason it's probably best not to have a set batch size as well, but I kept it in to ensure that the module could actually meet the specs.

# What I would do next...

This is an (unordered) list of how I would continue implementing this app.

## Acceptance Tests 

I use the term acceptance test instead of E2E test because it encompasses a wider range of automated tests. For this project I would probably include a smoke tests, performance test and behaviour test in the acceptance test phase. Each one of those would be a standalone binary that can test against the system in a production like environment. These three sets of tests would make up the acceptance test phase of a first draft delivery pipeline.  

## Externals

For the two external components (DB and message broker) I would create a small anticorruption layer (See Domain Driven Design) that translated from the external types to the types of our problem. For example, a PubSub anticorruption layer would include one function that transformed a channel of GCP style messages into a channel of Collector style messages.

## Monitoring 

The Collector can easily be modified using the middleware pattern to incorporate metrics. There are two types of metrics that would be useful to collect. First is system usage metrics, these can be fed back into performance acceptance tests so we can simluate the system at 5x peak usage. The second metrics are behaviour metrics which allow us to analyse the upstream service's log usage.

