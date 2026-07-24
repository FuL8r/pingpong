// package pingpong contains business logic

package pingpong

const AllowedBody = "ReallyNotBad"

type Decision int

const (
	Denied Decision = iota
	Allowed
)

type Command struct {
	NotBad bool
}

type Response struct {
	Decision Decision
	Body     string
}

// Service processes ping-pong type commands
type Service interface {
	Respond(Command) Response
}

// service is the default, stateless implementation of the rule
type service struct{}

// NewService returns the default ping-pong Service.
func NewService() Service {
	return service{}
}

// Respond a command carrying the NotBad
func (service) Respond(cmd Command) Response {
	if cmd.NotBad {
		return Response{Decision: Allowed, Body: AllowedBody}
	}
	return Response{Decision: Denied, Body: ""}
}
