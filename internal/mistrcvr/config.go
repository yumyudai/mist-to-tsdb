package mistrcvr

type Config struct {
	Mist struct {
		Endpoint		string	  `mapstructure:"endpoint"`
		Apikey			string	  `mapstructure:"apikey"`
		Debug			bool	  `mapstructure:"debug"`
	}                                         `mapstructure:"mist"`
	Tsdb struct {
		Enabled			bool	  `mapstructure:"enabled"`
		Driver			string	  `mapstructure:"driver"`
		Debug			bool	  `mapstructure:"debug"`
		BufSize			int	  `mapstructure:"channel_buffer_size"`
		Awstimestream		struct {
			Region		string	  `mapstructure:"aws_region"`
			Database	string	  `mapstructure:"database"`
			Maxretries	int	  `mapstructure:"max_retries"`
		}                                 `mapstructure:"aws_timestream"`
	}                                         `mapstructure:"tsdb"`
	Pubsub struct {
		Enabled			bool	  `mapstructure:"enabled"`
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
		Channel			string	  `mapstructure:"channel"`
		Datalayout		string	  `mapstructure:"data_layout"`
		Tsdb			struct {
			Table		string	  `mapstructure:"table"`
			Keys		[]string  `mapstructure:"keys"`
			Metrics		[]struct {
				Key	string	  `mapstructure:"key"`
				Type	string	  `mapstructure:"type"`
			}                         `mapstructure:"metrics"`
		}                                 `mapstructure:"tsdb"`
		Pubsub			struct {
			Topic		string	  `mapstructure:"topic"`
			Header		[]struct {
				Key	string	  `mapstructure:"key"`
				Value	string	  `mapstructure:"value"`
			}			  `mapstructure:"header"`
		}                                 `mapstructure:"pubsub"`
	}                                         `mapstructure:"datasource"`
}
