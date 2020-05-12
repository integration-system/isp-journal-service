package elastic

type Option func(rdc *RxElasticClient)

func WithInitializingErrorHandler(eh errorHandler) Option {
	return func(rdc *RxElasticClient) {
		rdc.eh = eh
	}
}

func WithInitializingHandler(handler initHandler) Option {
	return func(rdc *RxElasticClient) {
		rdc.initHandler = handler
	}
}
