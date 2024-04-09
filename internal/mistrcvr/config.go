package mistrcvr

type Config struct {
	Mist struct {
		Endpoint		string
		Apikey			string
		Debug			bool
		BufSize			int
	}
	Tsdb struct {
		Driver			string
		Debug			bool
		Awstimestream struct {
			Region		string
			Database	string
			Maxretries	int
		}
	}
	Datasource []struct {
		Channel			string
		Datalayout		string
		Table			string
		Keys			[]string
		Metrics []struct {
			Key		string
			Type		string
		}
	}
}
