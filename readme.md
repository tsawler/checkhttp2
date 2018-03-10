# checkhttp2

A simple plugin to check https urls, including http/2, and SSL expiration.

This pluig is meant to be used with [Nagios](https://www.nagios.org/).

The SSL check code is based on [asyncrsc/ssl_scan](https://github.com/asyncsrc/ssl_scan).
Nagios messaging is based on [newrelic/go_nagios](https://github.com/newrelic/go_nagios).

Compile for Digital Ocean: 

~~~
env GOOS=linux GOARCH=amd64 go build -o checkhttp2 main.go
~~~


## Installation

For Nagios 4 on Ubuntu 16.04, assuming that you followed [
these instructions](https://www.digitalocean.com/community/tutorials/how-to-install-nagios-4-and-monitor-your-servers-on-ubuntu-16-04),
just place in `/usr/local/nagios/libexec`, and make sure the file is executable.


## Usage

Run the command from cli as follows:

~~~
checkhttp2 -host <hostname.com> [-protocol http:https (default)] [-port 80|443 (default)|xxx ] [-cert true|false (default)]
~~~

Example: to check status of www.google.com:

~~~
checkhttp2 -host www.google.com
~~~

Example: to check SSL expiration date for www.google.com:

~~~
checkhttp2 -host www.google.com -cert true
~~~


Example: to check SSL expiration of somesite.com on port 5666:

~~~
checkhttp2 -host www.somesite.com -port 5666
~~~

## Integration with Nagios 4

Add this to `/usr/local/nagios/objects/commands.cfg` to test **HTTP/2 status**:

~~~
define command {
   command_name    check_http2
   command_line    /usr/local/nagios/libexec/checkhttp2 -host $ARG1$
}
~~~


Add this to `/usr/local/nagios/objects/commands.cfg` to test **SSL expiration status**:

~~~
define command {
   command_name    check_ssl_expiry
   command_line    /usr/local/nagios/libexec/checkhttp2 -host $ARG1$ -cert true
}
~~~


In individual files in `/usr/local/nagios/etc/servers`:

~~~
define service{
        use                     generic-service
        host_name               www.somesite.com
        service_description     Check HTTP2
        check_command           check_http2!www.somesite.com
}

define service{
        use                     generic-service
        host_name               www.somesite.com
        service_description     Check SSL Expiry
        check_command           check_ssl_exiry!www.somesite.com
}
~~~