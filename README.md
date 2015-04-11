Nebulo
======
### What Is It?
This is a web app written in Go. It uses goquery to scrape data from http://aqicn.org/ and stores it in a PostgreSQL database which can be accessed by the API layer to return a simple JSON response.

Will love to gather some feedback on how to improve this, so feel free to submit an issue and I'll be happy to learn!

### Getting Started
Clone the repository:

`
git clone git@github.com:jurvis/nebulo-backend.git
`

cd into your project directory and run the following to set up your GOPATH

`
export GOPATH=$(pwd)
`

To run:
` go run monitor.go `