## ipfinder

IP Finder tool, ipfinder collects ip address from different sources like Shodan, Zoomeye, Viewdns, dig command, etc.

## Installation
```
go install github.com/rix4uni/ipfinder@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/ipfinder/releases/download/v0.0.1/ipfinder-linux-amd64-0.0.1.tgz
tar -xvzf ipfinder-linux-amd64-0.0.1.tgz
rm -rf ipfinder-linux-amd64-0.0.1.tgz
mv ipfinder ~/go/bin/ipfinder
```
Or download [binary release](https://github.com/rix4uni/ipfinder/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/ipfinder.git
cd ipfinder; go install
```

## Usage
```
Usage:
  ipfinder [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  dig         Run the dig command to get DNS A records
  help        Help about any command
  shodan      Search for SSL details on Shodan (Website supports subdomains but recommended to use domain)
  viewdns     Fetch IP history for a domain (Website not supports subdomains)
  zoomeye     Fetch IP history for a domain (Website not supports subdomains)

Flags:
  -h, --help     help for ipfinder
  -t, --toggle   Help message for toggle
```

## TODO
```
add asn:AS51602 and org:"Dell Inc." features user can use like this
echo "asn:AS51602" | go run main.go shodan
```