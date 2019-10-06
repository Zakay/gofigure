# gofigure

Package gofigure is a fully functional configuration file parser with bells on

## Getting Started

Want to get started with this simpler way of configuring software?
Good. You shall now write .fig files for all of your projects.

## Example
A quick example of what you can do with gofigure
```
value: "string" ; Comment

integerKey: 42 # Comment

floatKey: 4.2

keyFromIdentifier: value

mapKey: {
  key: "string"
  int: integerKey
  array: [
    "Value"
    floatKey
  ]
}

arrayKey: [
  value
  value
  {
    key1: "test"
    key2: "test2"
    key3: ["test3" "test4"]
  }
  [
    value 
  ]
]
```

Should output the JSON

```
{
  "arrayKey": [
    "string", 
    "string", 
    {
      "key1": "test",
      "key2": "test2",
      "key3": ["test3", "test4"]
    },
    ["string"]
  ],
  "floatKey": 4.2,
  "integerKey": 42,
  "keyFromIdentifier": "string",
  "mapKey": {
    "array": ["Value", 4.2],
    "int": 42,
    "key": "string"
  },
  "value": "string"
}
```

For more advanced examples, look at the files in the [files](files) folder or check out the [EXAMPLES](EXAMPLES.md) file.

### Prerequisites

gofigure can be run 
* in a docker container
* locally on your machine - requires an installation of `golang` (this implementation was written in version 1.10)

### Installing

#### Linux
```
# Get the latest release and download
curl -O $(curl -s https://api.github.com/repos/techbuddyab/gofigure/releases/latest | grep 'browser_' | cut -d\" -f4)
chmod +x ./gofigure
mv ./gofigure /usr/local/bin
```

#### Building
```
git clone git@github.com:techbuddyab/gofigure
```

Now, you can choose to build and run the software locally or with docker.

#### Local build
```
# build with local go installation - binary in project root
make build-local
# run with example file
make run-local
```

Upon which you will have a built binary in the project root folder.

#### Docker build
The docker way has the same end result as the local build

```
# build in docker - binary in project root
make build
# run with example file
make run
```

#### After building
Try running the software with your own .fig files.

```
# Print JSON to stdout
./gofigure -i config.fig

# Write JSON to file
./gofigure -i config.fig -o config.json
```

## Running the tests

Running the tests locally is easy. Just `go test` it!

## Built With

* [participle](https://github.com/alecthomas/participle) - The parser library used
* No other dependencies - and we aim to keep it that way.

## Contributing

No `CONTRIBUTING.md` yet, but feel free to submit an issue if you have problems or ideas. PRs are also welcome, please submit a short description of your work and we'll take it up together from there.

## Authors

* **Zakay Danial** - *Idea, initial structure and mastermind* - [Zakay](https://github.com/Zakay)
* **Stefano De Colli** - *Counseling* - [sealos](https://github.com/sealos)
* **Simon Haak** - *Full implementation* - [shellkjell](https://github.com/shellkjell)

## License

This project is licensed under the Apache v2.0 License - see the [LICENSE](LICENSE) file for details

## Acknowledgments

* Examples from the alecthomas/participle repo were helpful - A truly dead simple parser.

