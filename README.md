BatSignal
=========

The `batsignal` program is designed to allow a user to use a personally-owned (sub) domain as a dynamic-dns pointed toward their (presumably) home IP.

This program is designed towards the particulars of the author's own setup, specifically: Linode as the DNS manager.

If you have a different DNS host, then you will have to modify the API calls to match that of your host (assuming they have an API).


Installation
-------------

1. Copy `batsignal.conf.example` to `batsignal.conf`. Modify the values to suit your environment and needs.
1. `pip install -r requirements.txt`

Then run `./batsignal`

License
-------

This program is licensed under the MIT License. See `LICENSE` for more
