$(function(){



    $( "#search_box" ).blur(function(){

        if($("#search_res:hover").length != 0 || $("#search_box:hover").length != 0){

            return

        }

        $("#search_box" ).val("")
        $("#search_res").remove();

    });

    $( "#search_box" ).keyup(function() {

       if(this.value.length<3) {
           return;
       }

       $.get("/search?q="+this.value,function(data){


            if(data.Code==200){

                $("#search_res").remove();
                $("#search_box").parent().append("<div style='min-width:250px;position:absolute;padding:5px;top:40px;background:#FFF;' id='search_res'></div>");


                for (var ind in data.Payload) {

                    obj = data.Payload[ind]
                    div = $("<div>").html("<a href='"+obj.Target+"' >"+obj.Label+"</a>");
                    $("#search_res").append(div);

                }



            }

       },"json");
    });



});