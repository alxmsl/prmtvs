package plexus

import "errors"

var (
	// ErrorCloseClosedPlexus defines error for a case when a closed Plexus is closed once again.
	ErrorCloseClosedPlexus = errors.New("closed the closed plexus")
	// ErrorNotSelectable defines error for a case when there is a try detected to use selectable functions on
	// a not selectable Plexus. This is denied because it can produce a deadlock.
	ErrorNotSelectable = errors.New("not selectable plexus")
	// ErrorSendToClosedPlexus defines error for a case when a send operation is detected for a closed Plexus.
	ErrorSendToClosedPlexus = errors.New("send to the closed plexus")
	// ErrorValueIsNotMergeable defines error for a case when sender sends non mergeable value with multiple
	// simultaneous senders. Plexus can not define which value should be passed to receivers in this case.
	ErrorValueIsNotMergeable = errors.New("value does not implement plexus.Mergeable")
	// ErrorUnknownState defines error for a case when Plexus has an incorrect state of senders/receivers.
	ErrorUnknownState = errors.New("plexus is in unknown state")
)

var (
	// ErrorQueueAlreadyExists defines error for a case of adding a queue with the same name.
	ErrorQueueAlreadyExists = errors.New("queue already exists")
	// ErrorQueueDoesNotExist defines error for a case of getting queue which has not been added.
	ErrorQueueDoesNotExist = errors.New("queue does not exist")
	// ErrorQueuesIsFull defines error for a case of adding a queue when there is no space for it.
	ErrorQueuesIsFull = errors.New("queues is full")
	// ErrorQueuesIsNotDefined defines error for a case of getting queue when queue is not fulfilled.
	ErrorQueuesIsNotDefined = errors.New("queues is not defined")
)
