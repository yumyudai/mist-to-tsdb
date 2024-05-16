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
		Awstimestream struct {
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
			Bootstrapsvrs	string	  `mapstructure:"bootstrap_servers"`
			Clientid	string	  `mapstructure:"client_id"`
			CidUseHostname	bool	  `mapstructure:"client_id_use_hostname"`
			Acks		string	  `mapstructure:"acks"`
		}                                 `mapstructure:"kafka"`
	}                                         `mapstructure:"pubsub"`
	Datasource []struct {
		Channel			string	  `mapstructure:"channel"`
		Datalayout		string	  `mapstructure:"data_layout"`
		Tsdb			struct {
			Table		string	  `mapstructure:"table"`
			Keys		[]string  `mapstructure:"keys"`
			Metrics []struct {
				Key	string	  `mapstructure:"key"`
				Type	string	  `mapstructure:"type"`
			}                         `mapstructure:"metrics"`
		}                                 `mapstructure:"tsdb"`
		Pubsub			struct {
			Topic		string	  `mapstructure:"topic"`
		}                                 `mapstructure:"pubsub"`
	}                                         `mapstructure:"datasource"`
}
