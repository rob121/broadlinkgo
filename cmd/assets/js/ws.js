
//ws = new WS("ws://example.com")
//ws.Open()
//ws.Close()
//ws.Send(msg)

var WS = function (url) {

    this.url = url;

    this.forceclose = false;

    this.reconnecting = false

    this.state = WebSocket.CONNECTING;

    this.wshandle = new WebSocket(this.url);

    this.handlers = [];

    this.AddListeners();

    this.retryperiod = 5000;

    par = this

    setInterval(function(){

        par.CheckConnection()

    },this.retryperiod);

}

WS.prototype.Handler = function(type,event){

    var fn = this.handlers[type];
    if (typeof fn === "function") {

        fn(event)
    }

}



WS.prototype.CheckConnection = function(){

    console.debug("Checking Connection",this.wshandle.readyState)

    if( this.wshandle.readyState != 1) {


        this.Reconnect();

    }


}

WS.prototype.AddListeners = function(){

    par = this

    this.wshandle.addEventListener('open', function(event){
        par.state = par.wshandle.readyState
        par.Handler("open",event)
    });

    this.wshandle.addEventListener('close', function(event){
        par.state = par.wshandle.readyState
        par.Handler("close",event)
    });

    this.wshandle.addEventListener('message', function(event){

     //   console.log(event)
        par.state = par.wshandle.readyState
        par.Handler("message",event)
    });

    this.wshandle.addEventListener('error', function(event){
        par.state = par.wshandle.readyState
        par.Handler("error",event)
    });


}



//open/close/message/send
WS.prototype.On = function(type,fn) {

    this.handlers[type] = fn


    //reload listeners based on new addition
    this.AddListeners()

}

WS.prototype.Send = function(msg){

    this.wshandle.send(msg)
}

WS.prototype.Reconnect = function () {

    if(this.reconnecting==true){

        return
    }

    this.reconnecting =  true

    this.wshandle = new WebSocket(this.url);

    this.AddListeners()

    this.reconnecting =  false



}

WS.prototype.Close = function () {
    this.forceclose = true
    this.wshandle.close()

}

