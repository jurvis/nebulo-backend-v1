Nebulo
======
### What Is It?
This is my first (tiny) web app written in Go. It uses goquery to scrape data from http://aqicn.org/ and stores is in a gkvlite persistent key-value store which can be accessed by the API layer to return a simple JSON response.

Will love to gather some feedback on how to improve this, so feel free to submit an issue and I'll be happy to learn!

### Getting Started
Clone the repository:

`
git clone git@github.com:jurvis/HazeSG.git
`

cd into your project directory and run the following to set up your GOPATH

`
source env.sh
`

To run:
` go run server.go `

### Brief API Docs
#### GET
> `/get` returns the current air quality in Singapore
```json
{
  "status": "Moderate.",
    "weather": {
    "PM25": "78",
    "PSI": "66",
    "Temp": "29"
  
    }

}
```

### POST
1. `/post` requires a json body, below is a sample:
```json
{
  "UUID": "hello world",
  "DeviceType": "iOS"

}
```
