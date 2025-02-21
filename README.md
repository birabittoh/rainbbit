# Rainbbit
A lightweight and self-hosted weather service.

![Desktop screenshot](/screenshots/desktop.png "Desktop")

Rainbbit does the following:
* reads free weather data from [OpenWeatherMap](https://openweathermap.org/);
* saves said data into a local SQLite database;
* plots several useful info;
* displays said plots without the need for JavaScript.

All of this is done in a single binary file that weighs less than 30MB.

## TODO
- [x] Add a light theme.
- [x] Cache the database requests.
- [x] Cache the plots.

## API endpoints

### GET /api/records
Retrieves weather records stored in the database.

### GET /api/latest
Gets the latest weather record.

### GET /api/conditions
Fetches all possible weather conditions.

### GET /api/meta
Provides the zone name, lists plottable measures and available themes.

### GET /api/plot/{measure}
Generates an SVG plot for any measure.

### GET /api/temp
Generates a custom SVG plot for temperatures.

## Instructions
First of all, create your own `.env` file:
```sh
cp .env.example .env
```

You should set your latitude and longitude, as well as a free OpenWeather 2.5 [API key](https://home.openweathermap.org/api_keys).

You can then either start the service locally:
```sh
go run .
```

Or use Docker:
```sh
docker compose up -d
```

## Optional variables
 Name        | Default value
-------------|----------------
`OWM_CRON`   |`0 0/30 * * * *`
`APP_ADDRESS`|`:3000`

## License
Rainbbit is licensed under MIT.
