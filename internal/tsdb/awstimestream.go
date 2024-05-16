package tsdb

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"

	"golang.org/x/net/http2"

	"mist-to-tsdb/pkg/mistdatafmt"
)

type tsdbIntfAwsTS struct {
	debug		bool
	database	string
	dataOut		map[string]tsdbIntfAwsTsDataOut

	awsConfig	*aws.Config
	awsTransport	*http.Transport
	awsSession	*session.Session
	tsWrite		*timestreamwrite.TimestreamWrite
}

type tsdbIntfAwsTsDataOutMetric struct {
	Key		string
	Type		string
}

type tsdbIntfAwsTsDataOut struct {
	Channel		string
	Table		string
	Keys		[]string
	Metrics		[]tsdbIntfAwsTsDataOutMetric
}

func tsdbIntfAwsTsNew(cfg TsdbIntfConf) (*tsdbIntfAwsTS, error) {
	var err error 

	// check valid parameter
	if cfg.DriverAwsTimeStream.Region == "" {
		return nil, fmt.Errorf("AWS region not specified")
	}

	if cfg.DriverAwsTimeStream.MaxRetries < 1 {
		return nil, fmt.Errorf("AWS max retries is below 1")
	}

	// start building interface
	r := &tsdbIntfAwsTS {
		debug:		cfg.Debug,
		database:	cfg.DriverAwsTimeStream.Database,
		dataOut:	make(map[string]tsdbIntfAwsTsDataOut),
	}

	// build data out
	for i := 0; i < len(cfg.Datasource); i++ {
		if cfg.Datasource[i].Channel == "" ||
			cfg.Datasource[i].Table == "" ||
			len(cfg.Datasource[i].Keys) < 1 {
			return nil, fmt.Errorf("Missing mandatory data out parameter")
		}

		dout := tsdbIntfAwsTsDataOut {
			Channel:	cfg.Datasource[i].Channel,
			Table:		cfg.Datasource[i].Table,
			Keys:		cfg.Datasource[i].Keys,
		}

		// check valid type
		// ref. https://docs.aws.amazon.com/timestream/latest/developerguide/writes.html
		for j := 0; j < len(cfg.Datasource[i].Metrics); j++ {
			t := strings.ToUpper(cfg.Datasource[i].Metrics[j].Type)
			switch t {
			// supported, pass through check
			case "BIGINT":
			case "BOOLEAN":
			case "DOUBLE":
			case "VARCHAR":
				
			// unsupported
			case "MULTI":
				return nil, fmt.Errorf("MULTI data type is not supported")
			default:
				return nil, fmt.Errorf("Unknown data type: %s", t)
			}

			m := tsdbIntfAwsTsDataOutMetric {
				Key:	cfg.Datasource[i].Metrics[j].Key,
				Type:	t,
			}

			dout.Metrics = append(dout.Metrics, m)
		}
		
		// finally add to map
		r.dataOut[cfg.Datasource[i].Channel] = dout
	}

	// aws connection
	// from https://github.com/awslabs/amazon-timestream-tools/blob/mainline/sample_apps/go/ingestion-csv-sample.go
	r.awsTransport = &http.Transport {
		ResponseHeaderTimeout: 20 * time.Second,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			DualStack: true,
			Timeout:   30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	http2.ConfigureTransport(r.awsTransport)

	r.awsConfig = &aws.Config {
		Region:				aws.String(cfg.DriverAwsTimeStream.Region),
		MaxRetries:			aws.Int(cfg.DriverAwsTimeStream.MaxRetries),
		HTTPClient:			&http.Client {Transport: r.awsTransport},
		CredentialsChainVerboseErrors:	aws.Bool(true),
	}

	err = r.initAws()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (i *tsdbIntfAwsTS) initConn() error {
	var err error 

	log.Printf("Connecting to AWS..")
	i.awsSession, err = session.NewSession(i.awsConfig)
	if err != nil {
		return err
	}

	i.tsWrite = timestreamwrite.New(i.awsSession)

	log.Printf("Connected to AWS..")
	return nil
}

func (i *tsdbIntfAwsTS) initAws() error {
	var err error

	// Get connection
	err = i.initConn()
	if err != nil {
		return err
	}

	// Get Database
	dbDescribe := &timestreamwrite.DescribeDatabaseInput {
		DatabaseName: aws.String(i.database),
	}

	dbOut, err := i.tsWrite.DescribeDatabase(dbDescribe)
	if err != nil {
		_, ok := err.(*timestreamwrite.ResourceNotFoundException)
		if ok {
			log.Printf("Database %s was not found, attempting to create..", i.database)
			dbCreate := &timestreamwrite.CreateDatabaseInput {
				DatabaseName: aws.String(i.database),
			}

			_, err = i.tsWrite.CreateDatabase(dbCreate)
			if err != nil {
				log.Printf("Could not create database: %v", err)
				return err
			}
		} else {
			log.Printf("Database check has returned error: %v", err)
			return err
		}
	} else {
		log.Printf("Got database: %v", dbOut)
	}

	// Describe table.
	for _, v := range i.dataOut {
		tblDescribe := &timestreamwrite.DescribeTableInput {
			DatabaseName: aws.String(i.database),
			TableName:    aws.String(v.Table),
		}

		tblOut, err := i.tsWrite.DescribeTable(tblDescribe)
		if err != nil {
			_, ok := err.(*timestreamwrite.ResourceNotFoundException)
			if ok {
				log.Printf("Table %s was not found, attempting to create..", v.Table)
				tblCreate := &timestreamwrite.CreateTableInput {
					DatabaseName: aws.String(i.database),
					TableName:    aws.String(v.Table),
				}
				_, err = i.tsWrite.CreateTable(tblCreate)

				if err != nil {
					log.Printf("Could not create table: %v", err)
					return err
				}
			} else {
				log.Printf("Table check has returned error: %v", err)
				return err
			}
		} else {
			log.Printf("Got table: %v", tblOut)
		}
	}

	return nil
}

func (i *tsdbIntfAwsTS) AddRecordStatsClient(channel string, data mistdatafmt.WsMsgClientStat) error {
	outParams, ok := i.dataOut[channel]
	if !ok {
		return fmt.Errorf("Data out parameters for channel %s was not found", channel)
	}

	if i.debug {
		log.Printf("Start processing add record for client statistics")
	}

	// build data record
	var keyDimensions []*timestreamwrite.Dimension
	for _, v := range outParams.Keys {
		dval, err := data.GetJsonKeyValueAsStr(v)
		if err != nil {
			return fmt.Errorf("JSON key %s does not exist", v)
		} else if dval == "" {
			continue
		}

		d := &timestreamwrite.Dimension {
			Name:	aws.String(v),
			Value:	aws.String(dval),
		}

		keyDimensions = append(keyDimensions, d)
	}

	if len(keyDimensions) < 1 {
		return fmt.Errorf("Key dimentions was empty")
	}

	counter := int64(0)
	curTime := time.Now().Unix()
	var records []*timestreamwrite.Record
	for _, v := range outParams.Metrics {
		rval, err := data.GetJsonKeyValueAsStr(v.Key)
		if err != nil {
			return fmt.Errorf("JSON key %s does not exist", v.Key)
		} else if rval == "" {
			return fmt.Errorf("Measurement data is empty for JSON key %s", v.Key)
		} 

		r := &timestreamwrite.Record {
			Dimensions:		keyDimensions,
			MeasureName:		aws.String(v.Key),
			MeasureValue:		aws.String(rval),
			MeasureValueType:	aws.String(v.Type),
			Time:			aws.String(strconv.FormatInt(curTime, 10)),
			TimeUnit:		aws.String("SECONDS"),
		}

		records = append(records, r)
		counter++

		if i.debug {
			log.Printf("Record#%d: %v", counter, r)
		}

		if counter % 100 == 0 {
			err := i.writeRecords(outParams.Table, records)
			if err != nil {
				return err
			}

			records = make([]*timestreamwrite.Record, 0)
			if i.debug {
				log.Printf("Flushing record as 100 record limit reached, wrote %d records", counter)
			}

			counter = 0
		}
	}

	if counter > 0 {
		err := i.writeRecords(outParams.Table, records)
		if err != nil {
			return err
		}

		if i.debug {
			log.Printf("Wrote %d records", counter)
		}
	}

	return nil
}

func (i *tsdbIntfAwsTS) AddRecordRaw(channel string, data string) error {
	log.Printf("Data layout raw is not supported for AWS TimeStream")

	return nil
}


func (i *tsdbIntfAwsTS) writeRecords(tbl string, records []*timestreamwrite.Record) error {
	var err error 

	writeRecordsInput := &timestreamwrite.WriteRecordsInput {
		DatabaseName:	aws.String(i.database),
		TableName:	aws.String(tbl),
		Records:	records,
	}

	_, err = i.tsWrite.WriteRecords(writeRecordsInput)
	if err != nil {
		log.Printf("Failed to write record: %v", err)

		// try reconn
		err = i.initConn()
		if err != nil {
			log.Printf("Re-connect failed: %v", err)
			return err
		}

		_, err = i.tsWrite.WriteRecords(writeRecordsInput)
		if err != nil {
			log.Printf("Failed to write after re-connect: %v", err)
			return err
		}
	}

	return nil
}
