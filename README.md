BatSignal
=========

The `batsignal` program is a dynamic IP publisher. When run, it determines the machine's outbound IP, and adds it as a DNS A record to a configured domain.

This program contacts Linode as the DNS manager. If you have a different DNS host, then you will have to modify the API calls to match that of your host (assuming they have an API). Open a github issue for discussion on adding hosts, if you would like support for one.

Installation
-------------

1. Grab a pre-compiled binary from the Releases tab (or build yourself from source. See **Building** below).
2. Run `TOKEN=[abc123] batsignal -d [SUBDOMAIN.DOMAIN.COM]`, substituting in your own API token (for your domain manager) and the desired (sub) domain to publish the DNS record.


Building
--------

This is a `go` program, using go modules. If you have a recent Go installation (1.11+) then using `go build` should produce a working binary.

License
-------

This program is licensed under the MIT License. See `LICENSE` for more
