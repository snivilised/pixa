# 🧩 Design Notes

This document serves as a place to contain notes/ideas about decisions made in design. It is not meant to be taken as gospel, it really is just a dumping ground for brain-storming sessions. Usually, design notes are just written into github issues, but the problem with that is that issues are closed and those notes get lost in history and are thus difficult to re-reference.

## 🚀 Workflow Patterns

🔎 See [Control-Flow Patterns](http://www.workflowpatterns.com/patterns/control/)

Actually, the referenced site looks to be very good as it deals with scenarios that I have been wrestling with in various parts of my snivilised projects, but I have dealt with them on an adhoc manner, coming up with my own way of doing things rather that looking at established patterns. This can serve as a very good reference. A lot of these patterns are relevant for web workflows, but they can be adapted to be used in other contexts too.

## 🧭 DDD in Go

🔎 See [how to impl ddd in go](https://programmingpercy.tech/blog/how-to-domain-driven-design-ddd-golang/)

* in ddd, a value object has no behaviour it just holds some state (ie it is a placeholder for some information). it is considered immutable and is unidentifiable.

* an entity is identifiable (by an ID)

* so inside the traversal callback, we create a value-object that then allows us the create th path-finder and whatever else is needed

* functionality does not go inside entity objects, they should be inside an aggregate. an aggregate combines multiple entities to create a full object. to create an aggregate, you must identify which of the entities is the root entity. an aggregate can also contain value-objects

* aggregates do not contain the marshalling/persistence tags (json/xml) for each of its members. the aggregate is not supposed to be ab to define the formatting of its members.

the reasons are:

* aggregates are stored by repositories (see repository pattern). the repository (typically hidden behind an interface) is a compound object which will contain many aggregates of a type. you might have a repository for mongo-db, another for ms-sql etc... ~ dao. Personally, I would do this a bit different; I would prefer a single repository that could be configured with different data access providers.

* a service is a higher level of abstraction than the repository. the repository will contain just a single type of entity, but typically in a system, we need many repositories. a service will couple together these repositories.

* use functional composition as a helper for creating factories (configuration pattern)

* sample needs an execution chain of objects to complete work, because multiple executions of magick with varying parameters will be required. In the full run, this chain is a chain of 1.

## Prompting for input in Command line

See: [promptui](https://github.com/manifoldco/promptui)

In resume scenarios, we need to be able to show a menu of the resume files found and let the user select which to use, using the up/down keys.

## sync.Pool

🔎 See: [Think In sync.pool](https://www.sobyte.net/post/2022-03/think-in-sync-pool/), [Martin Fowler - Registry](https://martinfowler.com/eaaCatalog/registry.html)

For each TraverseItem to be processed, we need a way to enter processing. We could create a new runner for every item. However, during long batch runs, that would mean creating and releasing many objects, placing more pressure on the GC. To avoid this, we can employ sync.pool, and let that allocate runners. In fact, we create a runner registry that wraps the object pool and we interact with the registry instead.

Ordinarily, the Registry would be contained within a service as the service would play host to many registries. But in Pixa, we only have a single registry, so it is not worth the overhead to introduce a service; doing so would be for the sake of it.
