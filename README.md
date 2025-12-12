## ipfinder

A fast and beginner-friendly command-line tool that extracts IP addresses and other facets from Shodan search queries. Supports both advanced Shodan query syntax and simple domain-based filtering.

## Installation
```
go install github.com/rix4uni/ipfinder@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/ipfinder/releases/download/v0.0.4/ipfinder-linux-amd64-0.0.4.tgz
tar -xvzf ipfinder-linux-amd64-0.0.4.tgz
rm -rf ipfinder-linux-amd64-0.0.4.tgz
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
- `--filter string`: Filter type (e.g., ssl, hostname, etc.) (default: "ssl")
- `-d, --delay int`: Delay between city queries in seconds (default: 0)
- `--retries int`: Maximum number of retries for 400/429 errors (default: 4, -1 for unlimited)
- `--source`: Include the source query in the output
- `--silent`: Suppress banner output
- `--version`: Print the version of the tool and exit
- `--verbose`: Enable verbose output (shows HTTP requests with status codes)

## Usage Examples

### Basic Usage
```yaml
# Using full Shodan query format (advanced)
echo 'ssl:"nvidia.com"' | ipfinder --silent
echo 'hostname:"sqrx.com"' | ipfinder --silent
echo 'ssl.cert.subject.cn:"sqrx.com"' | ipfinder --silent
echo 'org:"FIDELITY NATIONAL INFORMATION SERVICES"' | ipfinder --silent
echo 'asn:"AS3614"' | ipfinder --silent
echo 'ip:"173.0.84.0/24"' | ipfinder --silent
echo 'http.favicon.hash:"816615992"' | ipfinder --silent
cat subs.txt | ipfinder --silent

# Using --filter flag (beginner-friendly)
echo "nvidia.com" | ipfinder --silent --filter ssl
echo "sqrx.com" | ipfinder --silent --filter hostname
echo "example.com" | ipfinder --silent --filter ssl.cert.subject.cn
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
# Shows HTTP requests with status codes:
Running: [200] https://www.shodan.io/search/facet?query=ssl:"nvidia.com"&facet=ip
Running: [200] https://www.shodan.io/search/facet?query=ssl:"nvidia.com"&facet=city
Running: [200] https://www.shodan.io/search/facet?query=ssl:"nvidia.com"+city:"Boardman"&facet=ip
# ... followed by IP addresses

# On rate limiting (400/429 errors), shows retry attempts:
Running: [429] https://www.shodan.io/search/facet?query=ssl:"example.com"+city:"São+Paulo"&facet=ip
Running: [5s] https://www.shodan.io/search/facet?query=ssl:"example.com"+city:"São+Paulo"&facet=ip
Running: [429] https://www.shodan.io/search/facet?query=ssl:"example.com"+city:"São+Paulo"&facet=ip
Running: [10s] https://www.shodan.io/search/facet?query=ssl:"example.com"+city:"São+Paulo"&facet=ip
Running: [200] https://www.shodan.io/search/facet?query=ssl:"example.com"+city:"São+Paulo"&facet=ip
```

### Sequential Processing with Delay
For queries with many results (>= 1000), the tool automatically splits by city and processes them sequentially to avoid Cloudflare captcha:
```yaml
echo 'ssl:"nvidia.com"' | ipfinder --silent --delay 3
# Processes city queries sequentially with 3 second delay between each query
```

### Different Facets
```yaml
echo 'ssl:"example.com"' | ipfinder --silent --facet domain
# Returns domains instead of IPs

echo "example.com" | ipfinder --silent --filter ssl --facet domain
# Beginner-friendly way to get domains
```

### Retry Logic
The tool automatically retries on HTTP 400/429 errors with exponential backoff:
```yaml
echo "example.com" | ipfinder --silent --retries 5
# Retries up to 5 times on 400/429 errors (default: 4)

echo "example.com" | ipfinder --silent --retries -1
# Retries indefinitely until success
```

## Notes

- The tool supports subdomains but it's recommended to use domain names
- For more Shodan filters, see: https://www.shodan.io/search/filters
- After getting IPs, you can run naabu to get ports: `cat ips.txt | naabu -duc -silent -passive`
- Results are printed immediately as they're found (no buffering)
- When verbose mode is enabled, HTTP requests with status codes are shown on stderr, results on stdout
- The tool uses native Go HTTP requests (no external dependencies like curl)
- Automatic retry logic with exponential backoff (5s, 10s, 20s, 40s...) handles HTTP 400/429 errors
- Cities are processed sequentially to avoid Cloudflare captcha protection
- Special characters in city names (e.g., "São Paulo") are properly URL encoded