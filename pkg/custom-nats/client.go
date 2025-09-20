package custom_nats

type Client interface {
	Register(Router)
}
