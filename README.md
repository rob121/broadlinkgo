# Broadlinkgo
Broadlink golang http server

## Overview

This uses the 	"github.com/kwkoo/broadlinkrm" library

## Getting Started

Start the program 

go run maing.go --port=8000 --cmdpath=./


The program will look for devices on the network and then once found start the server listening on the port.

### Learning

Learning is super simple, just go to /learn and you'll get prompted for what to do. You will need to create a "commands" dir where --cmdpath points to

# Api

After learning codes the commands to use the api are dead simple

# /cmd/{name}

Sends a command using the name from learn, so if you did tv_on in learn then you would trigger it by using 

```GET /cmd/tv_on```

# /cmd/{name}:repeatn

For some buttons like volume, you will want to repeat them so you can use :N to repeat N times
```GET /cmd/tv_volup:5```
This will repeat the tv_volup command 5 times

# /macro/{name}:repeat/{name}:repeat

You may also want to turn up the volume and say change source, so do that with the macro end point

````GET /macro/tv_power/tv_volup:5/tv_source/tv_source_hdmi1```

There is also a magic "delay" command (1s) that will put in a delay for devices that can't handle the default 10ms break

````GET /macro/tv_power/tv_volup:5/tv_source/delay/delay/tv_source_hdmi1```

Add a 2 second delay to the macro


