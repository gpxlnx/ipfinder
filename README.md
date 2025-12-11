## ipfinder

IP Finder tool, ipfinder collects IP addresses from Shodan search queries.

## Installation
```
go install github.com/rix4uni/ipfinder@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/ipfinder/releases/download/v0.0.3/ipfinder-linux-amd64-0.0.3.tgz
tar -xvzf ipfinder-linux-amd64-0.0.3.tgz
rm -rf ipfinder-linux-amd64-0.0.3.tgz
mv ipfinder ~/go/bin/ipfinder
```
Or download [binary release](https://github.com/rix4uni/ipfinder/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/ipfinder.git
cd ipfinder; go install
```

## Usage

The tool reads Shodan queries from stdin and outputs IP addresses or other facets based on the query.

```
echo 'query' | ipfinder [flags]
cat queries.txt | ipfinder [flags]
```

### Flags

- `-f, --facet string`: Facet type (e.g., ip, domain, etc.) (default: "ip")
- `-c, --concurrent int`: Number of concurrent city queries (default: 1)
- `-d, --delay int`: Delay between batches of city queries in milliseconds (default: 0)
- `--source`: Include the source query in the output
- `--silent`: Suppress banner output
- `--version`: Print the version of the tool and exit
- `--verbose`: Enable verbose output (shows curl commands being executed)

## Usage Examples

### Basic Usage
```yaml
echo 'ssl:"nvidia.com"' | ipfinder --silent
echo 'hostname:"sqrx.com"' | ipfinder --silent
echo 'ssl.cert.subject.cn:"sqrx.com"' | ipfinder --silent
echo 'org:"FIDELITY NATIONAL INFORMATION SERVICES"' | ipfinder --silent
echo 'asn:"AS3614"' | ipfinder --silent
echo 'ip:"173.0.84.0/24"' | ipfinder --silent
echo 'http.favicon.hash:"816615992"' | ipfinder --silent
cat subs.txt | ipfinder --silent
```

### With Source Flag
```yaml
echo 'ssl:"dell.com"' | ipfinder --silent --source
# Output:
ssl:"dell.com"::192.168.1.1
ssl:"dell.com"::192.168.1.2
```

### With Verbose Flag
```yaml
echo 'ssl:"nvidia.com"' | ipfinder --silent --verbose
# Shows curl commands being executed:
Running: curl -s https://www.shodan.io/search/facet?query=ssl:"nvidia.com"&facet=ip
Running: curl -s https://www.shodan.io/search/facet?query=ssl:"nvidia.com"&facet=city
Running: curl -s https://www.shodan.io/search/facet?query=ssl:"nvidia.com"+city:"Boardman"&facet=ip
# ... followed by IP addresses
```

### Concurrent Processing
For queries with many results (>= 1000), the tool automatically splits by city and processes them concurrently:
```yaml
echo 'ssl:"nvidia.com"' | ipfinder --silent --concurrent 5 --delay 100
# Processes 5 city queries concurrently with 100ms delay between batches
```

### Different Facets
```yaml
echo 'ssl:"example.com"' | ipfinder --silent --facet domain
# Returns domains instead of IPs
```

## Notes

- The tool supports subdomains but it's recommended to use domain names
- For more Shodan filters, see: https://www.shodan.io/search/filters
- After getting IPs, you can run naabu to get ports: `cat ips.txt | naabu -duc -silent -passive`
- Results are printed immediately as they're found (no buffering)
- When verbose mode is enabled, curl commands are shown on stderr, results on stdout