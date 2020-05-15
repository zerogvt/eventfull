[![Build Status](https://travis-ci.org/zerogvt/eventfull.svg?branch=master)](https://travis-ci.org/github/zerogvt/eventfull)

# eventfull

A small application that can serve as a custom events source for [New Relic](https://docs.newrelic.com/docs/insights/insights-data-sources/custom-data/introduction-event-api).

# Interface

The interface consists of 2 json files:

## Event Template JSON
`event.json` is essentially the template of the custom event. As per NR requirements you need to include at least the field `eventType`. The rest of the fields are up to you.

Values can be set to go's template variables e.g. `{{.service_name}}` which you can define in the configuration json.

## Configuration JSON

`conf.json` holds up configuration values.
Here you can set specific values for templatized settings in the `event.json`. You can also set up specific values for your metrics, such as `metric_slo` (which needs to be a percentage) and `metric_cutoff_value`. Eventfull will produce events whose value will -over time- be below the cutoff value at `metric_cutoff_value` percentage. E.g. if you set `metric_cutoff_value` to 90.0 then 90.0% of the produced events will be below cutoff value.

## How to build and use
- clone this repo
- Edit `event.json` and `conf.json` to your specs.
- `$ go build`
- `$ ./eventfull`
  
## Use cases
- Test various custom events formats.
- Test SLO accounting by providing a "service" with guaranteed SLI metrics.
