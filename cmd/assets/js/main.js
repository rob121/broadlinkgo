
$(function(){

    $(document).on("click",".copy",function(){

        var ele = $(this).attr("copy-target");

        tel = document.createElement('textarea');
        tel.value = $(ele).text();
        document.body.appendChild(tel);
        tel.select();
        document.execCommand('copy');
        document.body.removeChild(tel);

    });

    $(document).on("click","#sidebarToggle",function(){



        if($("body").hasClass("sb-sidenav-toggled")){


            localStorage.setItem('menustate',"closed");

        }else{

            localStorage.setItem('menustate',"open");

        }


    });

    ms = localStorage.getItem('menustate');


    if(ms=="closed"){

       $("body").addClass("sb-sidenav-toggled");

    }else{

        $("body").removeClass("sb-sidenav-toggled");

    }



});


function reorder(){

    $("#checks > div").sort(function (a, b) {

        c = $(a).attr("priority");
        d = $(b).attr("priority");

        return parseInt(c) - parseInt(d);
    }).each(function () {
        var elem = $(this);
        elem.remove();
        $(elem).appendTo("#checks");
    });

}

function generateUUID() { // Public Domain/MIT
    var d = new Date().getTime();//Timestamp
    var d2 = (performance && performance.now && (performance.now()*1000)) || 0;//Time in microseconds since page-load or 0 if unsupported
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16;//random number between 0 and 16
        if(d > 0){//Use timestamp until depleted
            r = (d + r)%16 | 0;
            d = Math.floor(d/16);
        } else {//Use microseconds since page-load if supported
            r = (d2 + r)%16 | 0;
            d2 = Math.floor(d2/16);
        }
        return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
    });
}


function bindDeviceCheck(h){

    var id = "host_"+h.Host.replace(/[\W_]+/g,"");

    //ss = c.StateSince.replace("microseconds","* .0001").replace("minutes","* 60").replace("minute","* 60").replace("hours","* 3600").replace("hour","* 3600").replace("seconds","* 1").replace("second","* 1")

   // dur = eval(ss);


    //started = new Date().getTime() - dur;

    var elementExists = document.getElementById(id);

    html = renderHost(h);

    if(elementExists){

        //update
        $("#hosts").find("#"+id).html(html);


    }else{

        $("#hosts").append('<tr id="'+id+'">'+html+'</tr>');

    }

}

function renderTemplate(obj,tmpl){

    var template = document.getElementById(tmpl).innerHTML;

    var rendered = Mustache.render(template, obj);

    return rendered

}



function print(message) {
    console.log(message)
}
