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
- [ ] Cache the plots.
- [ ] Cache the database requests.

## Instructions
First of all, create your own `.env` file:
```sh
cp .env.example .env
```

You should set your required latitude and longitude, as well as a free OpenWeather 2.5 [API key](https://home.openweathermap.org/api_keys).

You can then either start the service locally:
```sh
go run .
```

Or use Docker:
```sh
docker compose up -d
```

## License
Rainbbit is licensed under MIT.
