package main

import (
	"fmt"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	Client *elastic.Client
)

type Config struct {
	Endpoints         []string
	Sniff             bool
	Gzip              bool
	EnableTrace       bool
	EnableHealthCheck bool
}

func MustInitESClient(config Config) *elastic.Client {
	var (
		err       error
		version   string
		settings  = make([]elastic.ClientOptionFunc, 0)
		endpoints = make([]string, 0)
		okNodes   = make([]string, 0)
		errNodes  = make([]string, 0)
	)

	if len(config.Endpoints) == 0 {
		logrus.Panic("must specify elastic endpoints")
	}

	for idx := range config.Endpoints {
		endpoints = append(endpoints, fmt.Sprintf("http://%s", config.Endpoints[idx]))
	}

	for idx := range endpoints {
		settings = append(settings, elastic.SetURL(endpoints[idx]))
	}

	if config.Sniff {
		settings = append(settings, elastic.SetSniff(true))
	}

	if config.Gzip {
		settings = append(settings, elastic.SetGzip(true))
	}

	if config.EnableTrace {
		settings = append(settings, elastic.SetTraceLog(logrus.New().WithField("trace", "elastic_trace")))
	}

	if !config.EnableHealthCheck {
		settings = append(settings, elastic.SetHealthcheck(config.EnableHealthCheck))
	} else {
		settings = append(settings, elastic.SetHealthcheckInterval(60*time.Second))
		settings = append(settings, elastic.SetHealthcheckTimeout(5*time.Second))
	}

	if Client, err = elastic.NewClient(settings...); err != nil {
		logrus.Panicf("init es client err: %v", err)
	}

	for idx := range endpoints {
		if version, err = Client.ElasticsearchVersion(endpoints[idx]); err != nil {
			errNodes = append(errNodes, endpoints[idx])
		} else {
			okNodes = append(okNodes, endpoints[idx])
		}
	}

	switch len(okNodes) {
	case 0:
		logrus.Panicf("all nodes: %+v unavailable", endpoints)
	case len(config.Endpoints):
		logrus.Infof("connect to elastic[version: %s] success, all nodes: %+v available", version, endpoints)
	default:
		logrus.Warnf("connect to elastic all nodes: %+v, err nodes: %v", endpoints, errNodes)
	}

	return Client
}
