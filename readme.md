# checkHttp2

A simple plugin to check https urls, including http/2.

Used with [Nagios](https://www.nagios.org/).

Compile for Digital Ocean: `env GOOS=linux GOARCH=amd64 go build -o chechHttp2 main.go`

## Usage

For Nagios 4 on Ubuntu 16.04, assuming that you followed [
these instructions](https://www.digitalocean.com/community/tutorials/how-to-install-nagios-4-and-monitor-your-servers-on-ubuntu-16-04),
just place in `/usr/local/nagios/libexec`, and make sure the file is executable.

~~~
checkHttp2 <url>
~~~

Note that https:// is added to the URL automatically, *so don't include it*.


Add this to `/usr/local/nagios/objects/commands.cfg`:

~~~
define command {
   command_name    check_http2
   command_line    /usr/local/nagios/libexec/checkHttp2 $ARG1$
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
~~~