package main

func main() {
	config, err := initConfig()
	if err != nil {
		panic(err)
	}
	stores, err := initStores(config)
	if err != nil {
		panic(err)
	}

	service := &WalletService{
		stores: stores,
	}

	if err := service.ServeGRPC(); err != nil {
		panic(err)
	}
}
