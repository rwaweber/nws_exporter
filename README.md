# nws_exporter #

Prometheus exporter for the national weather service rest api

Documentation can be found
[here](https://www.weather.gov/documentation/services-web-api).

The endpoints refresh around once every hour or two.
It doesn't seem like the interval that it gets updated at is too consistent.

# Metrics supported
| name | unit | type |
|--------------|---------|-------|
| nws_humidity | percent | guage |
