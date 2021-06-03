# Broadlinkgo

Broadlink golang http server api

## Overview

This api allows access to the Broadlink family of devices for IR/RM control via an http api.  There is a web portal for basic management as well.

![Image of Home](broadlink_home.png?raw=true)



## Getting Started

Start the program 

```
broadlinkgo --port=8000 --cmdpath=./ 
```

Cmdpath is the location where a "commands" folder exists (it will be created if not present) this is where learned codes will be placed.

For linux/systemd systems a service file is included, the default command dir is "/etc/broadlinkgo" and the systemd unit file is looking for the binary in /usr/local/bin


The program will look for devices on the network and then once found start the server listening on the port. It will continually scan for devices so if more are added later they will be auto added without needing to re-start

## Upgrading from v1

To upgrade from v1 start the program with these flags (only necessary the first time)

```
broadlinkgo --upgrade=true

```

This will import your ir codes to the new format, your api commands will stay the same so unless you choose to re-map things, everything will continue working.

## Building from source

```
cd cmd
go build -o broadlinkgo
```

This will package up everything into a binary "broadlinkgo". I've included a few popular build types binaries in releases.


## Web Interface

This has received an overhaul in v2. See sections below

### Devices

Devices are listed and can be edited with descriptive labels. On startup broadlinkgo will searchyour network and mark a check next to devices it's found. Not finding the device is not necessarily an issue depending on our network setup

### Commands

Commands allow you to manage a command, set an icon and color for the button related to the command.

### Macros

New in v2 macros let you compose commands and save them as a macro for later reference.

### Equipment

New in v2 equipment is a helpful description of your target devices, tv's, stereos, etc.

## Remotes

New in v2 List remotes and associate buttons, also a remote view to be used on mobile.

```
broadlinkgo --port=8000 --cmdpath=./ --mode=manual
```

A button will appear on the interface allowing you to add a device manually skipping the discovery step. Special care should be taken to make sure you pick the correct device type to connect to when using this mode.

### Learning

Learning is super simple, just go to /learn via the home page and you'll get prompted for what to do. 

![Image of Home](broadlink_learn.png?raw=true)


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
