package handlers

import (
	"time"

	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
)

type CommandHandlerFixture struct {
	*gunit.Fixture

	input   chan RequestMessage
	output  chan EventMessage
	router  *FakeLockerRouter
	handler *CommandHandler
}

///////////////////////////////////////////////////////////////

func (this *CommandHandlerFixture) Setup() {
	this.input = make(chan RequestMessage, 4)
	this.output = make(chan EventMessage, 4)
	this.router = &FakeLockerRouter{}
	this.handler = NewCommandHandler(this.input, this.output, this.router, this.router)
}

///////////////////////////////////////////////////////////////

func (this *CommandHandlerFixture) TestRouterHandlesInputMessages() {
	this.input <- RequestMessage{Message: 1}
	this.input <- RequestMessage{Message: 2}
	this.input <- RequestMessage{Message: 3}

	this.listen()

	this.So(this.router.handled, should.Resemble, []interface{}{1, 2, 3})
}

///////////////////////////////////////////////////////////////

func (this *CommandHandlerFixture) TestLockProperlyManaged() {
	this.input <- RequestMessage{Message: 1}
	this.input <- RequestMessage{Message: 2}
	this.input <- RequestMessage{Message: 3}

	this.listen()

	lockHandleUnlockSequence := []time.Time{this.router.locks[0], this.router.handles[0], this.router.unlocks[0]}
	this.So(lockHandleUnlockSequence, should.BeChronological)
	this.So(this.router.locked, should.Equal, 0)
	this.So(len(this.router.locks), should.Equal, 1)
	this.So(len(this.router.unlocks), should.Equal, 1)
}

///////////////////////////////////////////////////////////////

func (this *CommandHandlerFixture) TestAllResultingEventsPassedToNextPhase() {
	context1 := &FakeRequestContext{id: 1}
	context2 := &FakeRequestContext{id: 2}
	this.input <- RequestMessage{Message: 1, Context: context1}
	this.input <- RequestMessage{Message: 2, Context: context2}
	this.router.results = append(this.router.results, []interface{}{"1a", "1b"})
	this.router.results = append(this.router.results, []interface{}{"2a", "2b"})

	this.listen()

	this.So(<-this.output, should.Resemble, EventMessage{Message: "1a", Context: context1})
	this.So(<-this.output, should.Resemble, EventMessage{Message: "1b", Context: context1, EndOfBatch: true})
	this.So(<-this.output, should.Resemble, EventMessage{Message: "2a", Context: context2})
	this.So(<-this.output, should.Resemble, EventMessage{Message: "2b", Context: context2, EndOfBatch: true})
}

///////////////////////////////////////////////////////////////

func (this *CommandHandlerFixture) TestNoResultsPassesContextToNextPhase() {
	context1 := &FakeRequestContext{id: 1}
	this.input <- RequestMessage{Message: 1, Context: context1}

	this.listen()

	this.So(len(this.output), should.Equal, 1)
	this.So(<-this.output, should.Resemble, EventMessage{Context: context1, EndOfBatch: true})
}

func (this *CommandHandlerFixture) listen() {
	close(this.input)
	this.handler.Listen()
}

///////////////////////////////////////////////////////////////

type FakeLockerRouter struct {
	handled []interface{}
	results [][]interface{}
	locked  int
	handles []time.Time
	locks   []time.Time
	unlocks []time.Time
}

func (this *FakeLockerRouter) Handle(item interface{}) []interface{} {
	this.handled = append(this.handled, item)
	this.handles = append(this.handles, time.Now())

	if len(this.results) == 0 {
		return nil
	}

	return this.results[len(this.handled)-1]
}

func (this *FakeLockerRouter) Lock() {
	this.locked++
	this.locks = append(this.locks, time.Now())
}

func (this *FakeLockerRouter) Unlock() {
	this.locked--
	this.unlocks = append(this.unlocks, time.Now())
}