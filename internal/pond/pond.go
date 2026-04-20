package pond

// Pond composes both services. Temporary compatibility shim for existing tests.
type Pond struct {
	*AccountService
	*SubscriptionService
}

func New(fetcher FeedFetcher, sources Sources, subscriptions Subscriptions, users Users, credentials Credentials, sessions Sessions) *Pond {
	return &Pond{
		AccountService:      NewAccountService(users, credentials, sessions),
		SubscriptionService: NewSubscriptionService(fetcher, sources, subscriptions),
	}
}
