# covid
using [colly](https://github.com/gocolly/colly) obtain covid data from worldometers.info per day, and show a 'daily-cases/daily-deaths' chart using [d3](https://github.com/d3/d3)


## step-1: update files to server
or using some local web server like Python:
```
python -m http.server 80
```

## step-2: build and first run
```
go build
```
this generate a executable file `covid`.

then run:
```
./covid
```
this obtain all data from worldometers.info and save to `data` folder.

## step-3: add sceduled task
let `covid` auto obtain data once per day.

window: using scheduled tasks
linux: using crontab
```
crontab -e
15 8-10 * * *  /your_server/covid/covid
```

## step-4: visit site
```
https://your_server.com/covid
```
