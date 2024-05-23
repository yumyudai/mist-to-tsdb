package mistpoller

type Config struct {
	Mist struct {
		Endpoint		string	  `mapstructure:"endpoint"`
		Apikey			string	  `mapstructure:"apikey"`
		Debug			bool	  `mapstructure:"debug"`
	}                                         `mapstructure:"mist"`
	Pubsub struct {
		Driver			string	  `mapstructure:"driver"`
		Debug			bool	  `mapstructure:"debug"`
		BufSize			int	  `mapstructure:"channel_buffer_size"`
		Kafka struct {
			Async		bool	  `mapstructure:"async"`
			Bootstrapsvrs	string	  `mapstructure:"bootstrap_servers"`
			Clientid	string	  `mapstructure:"client_id"`
			CidUseHostname	bool	  `mapstructure:"client_id_use_hostname"`
			ClientOpts	[]struct {
				Key	string	  `mapstructure:"key"`
				Value	string	  `mapstructure:"value"`
			}			  `mapstructure:"client_options"`
			FlushWait	int	  `mapstructure:"flush_wait_seconds"`
		}                                 `mapstructure:"kafka"`
	}                                         `mapstructure:"pubsub"`
	Datasource []struct {
		Uri			string	  `mapstructure:"uri"`
		Datalayout		string	  `mapstructure:"data_layout"`
		Interval		int       `mapstructure:"interval"`
		WatchKeys		[]string  `mapstructure:"watch_change_keys"`
		UniqueKey		string    `mapstructure:"unique"`
		Pubsub			struct {
			Topic		string	  `mapstructure:"topic"`
			Header		[]struct {
				Key	string	  `mapstructure:"key"`
				Value	string	  `mapstructure:"value"`
			}			  `mapstructure:"header"`
		}                                 `mapstructure:"pubsub"`
	}                                         `mapstructure:"datasource"`
}
