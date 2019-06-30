# Broadlinkgo

Broadlink golang http server api

## Overview

This api allows access to the Broadlink family of devices for IR/RM control via an http api.

## Getting Started

Start the program 

```
go run maing.go --port=8000 --cmdpath=./ 
```

Cmdpath is the location where a "commands" folder exists (it will be created if not present) this is where learned codes will be placed.


The program will look for devices on the network and then once found start the server listening on the port. It will continually scan for devices so if more are added later they will be auto added without needing to re-start

### Home Page

This page gives overall status and ability to manage learned commands

### Learning

Learning is super simple, just go to /learn via the home page and you'll get prompted for what to do. 

# Api

After learning codes the commands to use the api are dead simple

## Sending a command

```
GET /cmd/{name}
```

Sends a command using the name from learn, so if you did tv_on in learn then you would trigger it by using 

```
GET /cmd/tv_on
```

## Sending a command that repeats

```
GET /cmd/{name}:repeatn
```

For some buttons like volume, you will want to repeat them so you can use :N to repeat N times

```
GET /cmd/tv_volup:5
```

This will repeat the tv_volup command 5 times

## Sending multiple commands as a macro

```
GET /macro/{name}:repeat/{name}:repeat
```

You may also want to turn up the volume and say change source, so do that with the macro end point

```
GET /macro/tv_power/tv_volup:5/tv_source/tv_source_hdmi1
```

There is also a magic "delay" command (1s) that will put in a delay for devices that can't handle the default 10ms break

```
GET /macro/tv_power/tv_volup:5/tv_source/delay/delay/tv_source_hdmi1
```

## Removing a command 

```
POST /remove/{name}
```

Removing the volup command for a vizio tv

```
POST /remove/vizio_volup
```


Add a 2 second delay to the macro

# Credit

This builds upon the work of a few other projects most notably https://github.com/kwkoo/golang_broadlink_rm
