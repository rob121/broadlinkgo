# Broadlinkgo

Broadlink golang http server api

## Overview

This api allows access to the Broadlink family of devices for IR/RM control via an http api.


![Image of Home](broadlink_home.jpg?raw=true)


## Getting Started

Start the program 

```
broadlinkgo --port=8000 --cmdpath=./ 
```

Cmdpath is the location where a "commands" folder exists (it will be created if not present) this is where learned codes will be placed.

For ease a systemd service file is included, the default command dir is "/etc/broadlinkgo" and the systemd unit file is looking for the binary in /usr/local/bin


The program will look for devices on the network and then once found start the server listening on the port. It will continually scan for devices so if more are added later they will be auto added without needing to re-start

## Building from source

The program uses rice (https://github.com/GeertJohan/go.rice) to embed html

Once you have the rice tool build:

```
cd cmd
rice embed-go
go build
```

This will package up everything into a binary called "cmd", just rename to broadlinkgo. I've included a few popular build types binaries in releases.


### Home Page

This page gives overall status and ability to manage learned commands. You can connect to it by visiting the ip and port of the machine the software is running on ie http:/127.0.0.1:8000 (default)

## Adding Devices

Add the Broadlink devices to your network per the manufacturer instructions, once the broadlinkgo software is started it will find them on the network and make them available in the web interface. 

### Learning

Learning is super simple, just go to /learn via the home page and you'll get prompted for what to do. 

![Image of Home](broadlink_learn.jpg?raw=true)

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

Add a 2 second delay in the macro

## App Status

```
GET /status
```

Returns json for devices and commands that are configured

## Sending a command to a specific device

You can send to a specific device by using the mac enumerated in the web page under devices

```
GET /device/{MAC}/cmd/....
GET /device/{MAC}/macro/....
```

For example, send tv_on to device 8f:47:4E:I9

```
GET /device/8f:47:4E:I9/cmd/tv_on
```


## Removing a command 

```
POST /remove/{name}
```

Removing the volup command for a vizio tv

```
POST /remove/vizio_volup
```


## Sending command to a specifc device

Using the mac address listed in devices on the web portal you can send the above commands to a specific device



# Credit

This builds upon the work of a few other projects most notably https://github.com/kwkoo/golang_broadlink_rm
