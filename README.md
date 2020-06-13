[![Build Status](https://travis-ci.org/zerogvt/eventfull.svg?branch=master)](https://travis-ci.org/github/zerogvt/eventfull) [![Go Report Card](https://goreportcard.com/badge/github.com/zerogvt/eventfull)](https://goreportcard.com/report/github.com/zerogvt/eventfull)

# Eventfull

Eventfull is a tiny application that can serve as a custom events source for [New Relic](https://docs.newrelic.com/docs/insights/insights-data-sources/custom-data/introduction-event-api).

# Preconditions
You need next environmental variables set:
`NEW_RELIC_ACCOUNT_ID` and `NEW_RELIC_INSIGHTS_KEY` as per [New Relic referrence](https://docs.newrelic.com/docs/insights/insights-data-sources/custom-data/introduction-event-api).

# User Interface
The "user interface" consists of 2 json files:

## Event Template JSON
`event.json` is essentially the template of the custom event. As per NR requirements you need to include at least the field `eventType`. The rest of the fields are up to you.

Values can be set to go's template variables e.g. `{{.service_name}}` which you can define in the configuration json.

## Configuration JSON

`conf.json` holds up configuration values.
Here you can set specific values for templatized settings in the `event.json`. 
If you set a `url` then the events will be sent to that endpoint rather than NR.
You can set up specific values for your metrics, such as `sli` (which needs to be a percentage) and `metric_cutoff_value`. Eventfull will produce events whose value will -over time- be below the cutoff value at `sli` percentage. E.g. if you set `metric_cutoff_value` to 180 and `sli` to 90.0 then 90.0% of the produced events will be below 180 thus giving you a guaranteed Service Level Indicator of 90%. `slo` (Service Level Objective) on the other hand is the wanted objective for your service. You can start of next sample configuration if you want to use owr own server for experimentation:
```
{
    "url": "http://localhost:8080/ingest",
    "service": "test_service",
    "metric":  "test_metric",
    "cutoff_value": 100,
    "sli": 90,
    "slo": 99,
    "repeat_every_msecs": 1
}
```

## How to build and use
- clone this repo
- `cd eventfull\eventfull` 
- Edit `event.json` and `conf.json` to your specs.
- `$ go build`
- `$ ./eventfull client`
  
## Run test local server
- `$ ./eventfull server`

## Local server endpoints:
- `http://localhost:8080/stats`
  
## Use cases
- Test various custom events formats.
- Test SLO accounting by providing a "service" with guaranteed SLI metrics.
