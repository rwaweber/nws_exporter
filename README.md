# nws_exporter #

Prometheus exporter for the national weather service observation api

Documentation can be found
[here](https://www.weather.gov/documentation/services-web-api).

The endpoints refresh around once every hour or two.
It doesn't seem like the interval that it gets updated at is too consistent.

All we are doing here is making a GET request to this URL:

https://api.weather.gov/stations/<Station_Name>/observations/latest

A less than perfect way to find a station near you would be to go to
https://www.weather.gov and click through the map to find where you are,
and that should land you on a page leading with "Current conditions at
<City, Location> (Station Name)"

Once you've found the station name, thats all we need to get started. If we
found the station name to be KRKS an example run would look like:

```
nws_exporter -station KRKS
```

# Installation

```
git clone https://github.com/rwaweber/nws_exporter
cd nws_exporter
go build
```

After building, the `nws_exporter` executable can be found in the current
directory.

# Metrics supported
| name | unit | type |
|--------------|----------|-------|
| `nws_humidity` | percent  | guage |
| `nws_barometric_pressure` | pascals | guage |
| `nws_dewpoint` | celsius | guage |
| `nws_humidity` | percent | guage |
| `nws_temperature` | celsius | guage |
| `nws_visibility` | meters | guage |
| `nws_wind_direction` | degrees (angle) | guage |
| `nws_wind_speed` | kilometers per hour | guage |

# Usage
options:
```
Usage of nws_exporter:
  -addr string
        nws address (default "api.weather.gov")
  -backofftime int
        backofftime in seconds (default 100)
  -help
        help info
  -localaddr string
        The address to listen on for HTTP requests (default ":8080")
  -station string
        nws address (default "KPHL")
  -timeout int
        timeout in seconds (default 10)
  -verbose
        verbose logging
```
